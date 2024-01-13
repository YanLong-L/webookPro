package web

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webookpro/internal/domain"
	"webookpro/internal/service"
)

type UserHandler struct {
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	svc         *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		emailExp:    emailExp,
		passwordExp: passwordExp,
		svc:         svc,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginJWT)
	ug.GET("/profile", u.ProfileJWT)
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
		ctx.String(http.StatusOK, "邮箱重复，请换一个邮箱")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
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
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 生成 jwt token
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		Uid:       user.Id,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)), // 暂时设置成1分钟过期
		},
	})
	tokenStr, _ := tokenObj.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	ctx.Header("x-jwt-token", tokenStr)
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
	claims, ok := uc.(UserClaims)
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

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
