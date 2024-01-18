package web

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"webookpro/internal/service"
	"webookpro/internal/service/oauth2/wechat"
)

type OAuth2WechatHandler struct {
	svc     wechat.Oauth2Service
	userSvc service.UserService
	jwtHandler
	stateKey []byte
}

func NewOAuth2WechatHandler(svc wechat.Oauth2Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:      svc,
		userSvc:  userSvc,
		stateKey: []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf1"),
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthUrl)
	g.Any("/callback", h.CallBack)
}

// AuthUrl 构造跳转到微信扫码登录页面的url 并返回给前端跳转
func (h *OAuth2WechatHandler) AuthUrl(ctx *gin.Context) {
	// 生成一个state token 塞到 cookie 在微信的callback中校验
	state := uuid.New().String()
	url, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造微信登录URL失败",
		})
		return
	}
	err = h.SetStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Data: url,
		Msg:  "ok",
	})
}

// 微信扫码登录成功后，微信带着code和status回调到这里

func (h *OAuth2WechatHandler) CallBack(ctx *gin.Context) {
	// 获取微信回调回来的code和state
	code := ctx.Query("code")
	state := ctx.Query("state")
	// 校验state是否正确
	err := h.VerifyStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "state不一致",
		})
		return
	}
	// 校验code是否正确
	wechatInfo, err := h.svc.VerifyCode(ctx, code, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 查找或新建用户
	user, err := h.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = h.setJwtToken(ctx, user)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "OK",
	})
}

// SetStateCookie authurl请求到微信时，将state数据写到jwttoken并设置到cookie
func (h *OAuth2WechatHandler) SetStateCookie(ctx *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{
		State: state,
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return err
	}
	ctx.SetCookie("jwt-state",
		tokenStr, 600,
		"/oauth2/wechat/callback",
		"", false, false)
	return nil
}

// VerifyStateCookie 校验微信callback回来带的cookie是否和拿到的一致
func (h *OAuth2WechatHandler) VerifyStateCookie(ctx *gin.Context, state string) error {
	// 从cookie中拿token
	tokenStr, err := ctx.Cookie("jwt-state")
	if err != nil {
		return err
	}
	var claims StateClaims
	_, err = jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil {
		return err
	}
	// 校验 state 是否相等
	if state != claims.State {
		return errors.New("state不一致")
	}
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
