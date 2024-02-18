package web

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"webookpro/internal/domain"
	"webookpro/internal/errs"
	"webookpro/internal/service"
	ijwt "webookpro/internal/web/jwt"
)

const biz = "login"

var _ handler = (*UserHandler)(nil)

type UserHandler struct {
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	phoneExp    *regexp.Regexp
	svc         service.UserService
	codeSvc     service.CodeService
	ijwt.JwtHandler
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, jwtHdl ijwt.JwtHandler) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
		phoneRegexPattern    = "^1[0-9]{10}$"
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	phoneExp := regexp.MustCompile(phoneRegexPattern, regexp.None)
	return &UserHandler{
		emailExp:    emailExp,
		passwordExp: passwordExp,
		phoneExp:    phoneExp,
		svc:         svc,
		codeSvc:     codeSvc,
		JwtHandler:  jwtHdl,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginJWT)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)
	ug.POST("/refresh_token", u.RefreshToken)
	ug.POST("/logout", u.LogOutJwt)
}

// SignUp 用户注册
func (u *UserHandler) SignUp(ctx *gin.Context) {
	// 参数接收
	type SignUpRequest struct {
		Email           string `json:"email" binding:"required"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirmPassword" binding:"required"`
	}
	var req SignUpRequest
	err := ctx.Bind(&req)
	if err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	// 参数校验
	// 校验邮箱格式
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱格式错误")
		return
	}
	// 校验密码是否一致
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "密码不一致")
		return
	}
	// 校验密码格式
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	// 调用service
	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		// 这是复用
		span := trace.SpanFromContext(ctx.Request.Context())
		span.AddEvent("邮件冲突")
		ctx.String(http.StatusOK, "邮箱重复，请换一个邮箱")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

// SendLoginSMSCode 发送验证码
func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone" binding:"required"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
		})
		return
	}
	// 校验手机号格式
	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "输入有误",
		})
		return
	}
	// 发送验证码
	err = u.codeSvc.Send(ctx, biz, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "发送成功",
	})
}

// LoginSMS 用户登录 by 用户验证码
func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
		})
		return
	}
	// 校验手机号格式
	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "输入有误",
		})
		return
	}
	// 校验验证码
	ok, err = u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码错误",
		})
		return
	}
	// 校验通过，拿到user 对象
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 生成 jwt token access token 和 refresh token
	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "登录成功",
	})
	return
}

// LoginJWT 用户登录 by jwt token
func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type loginReq struct {
		Email    string
		Password string
	}
	var req loginReq
	// 参数绑定
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	// 登录
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: errs.UserInvalidOrPassword,
			Msg:  "用户不存在或密码错误",
		})
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 生成 jwt  access token 和 refresh token
	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

// Login 用户登录 by session
func (u *UserHandler) Login(ctx *gin.Context) {
	type loginReq struct {
		Email    string
		Password string
	}
	var req loginReq
	// 参数绑定
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	// 登录
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 登录成功后，将用户id写入session
	session := sessions.Default(ctx)
	session.Set("user_id", user.Id)
	session.Options(sessions.Options{
		MaxAge: 60, // 设置让它60s过期
	})
	err = session.Save()
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}

	ctx.String(http.StatusOK, "登录成功")
}

// LogOut 基于session的退出登录
func (u *UserHandler) LogOut(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Options(sessions.Options{
		MaxAge: -1,
	})
	if err := session.Save(); err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	ctx.String(http.StatusOK, "退出登录成功")
}

// ProfileJWT 用户信息
func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	uc, ok := ctx.Get("claims")
	if !ok {
		// 你可以考虑监控住这里
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	claims, ok := uc.(ijwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	userId := claims.Uid
	user, err := u.svc.Profile(ctx, userId)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, fmt.Sprintf("这是user:%#v 的profile", user))
}

// Profile 用户信息
func (u *UserHandler) Profile(ctx *gin.Context) {
	ctx.String(http.StatusOK, "这是一条默认的profile")
}

// RefreshToken 前端根据 refresh token来换取 access token
func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	// 只有这个接口，拿出来的才是 refresh_token，其它地方都是 access token
	refreshToken := u.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = u.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// 要么 redis 有问题，要么已经退出登录
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 搞个新的 access_token
	err = u.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "刷新成功",
	})
}

// 基于jwt的退出登录
func (u *UserHandler) LogOutJwt(ctx *gin.Context) {
	// 清除 accesstoken 和 freshtoken
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "退出登录失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "退出登录OK",
	})
}
