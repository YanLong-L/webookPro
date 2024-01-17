package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webookpro/internal/service"
	"webookpro/internal/service/oauth2/wechat"
)

type OAuth2WechatHandler struct {
	svc     wechat.Oauth2Service
	userSvc service.UserService
	jwtHandler
}

func NewOAuth2WechatHandler(svc wechat.Oauth2Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthUrl)
	g.Any("/callback", h.CallBack)
}

// AuthUrl 构造跳转到微信扫码登录页面的url 并返回给前端跳转
func (h *OAuth2WechatHandler) AuthUrl(ctx *gin.Context) {
	url, err := h.svc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造微信登录URL失败",
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
