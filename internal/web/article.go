package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webookpro/internal/domain"
	"webookpro/internal/service"
	ijwt "webookpro/internal/web/jwt"
	"webookpro/pkg/logger"
)

type handler interface {
	RegisterRoutes(server *gin.Engine)
}

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc service.ArticleServcie
	l   logger.Logger
}

func NewArticleHandler(svc service.ArticleServcie, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}

func (u *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/articles")
	ug.POST("/edit", u.Edit)
	ug.POST("/publish", u.Publish)
	ug.POST("/withdraw", u.Withdraw)
}

// Edit 创作者编辑一篇文章并保存
func (u *ArticleHandler) Edit(ctx *gin.Context) {
	// 参数接收 & 校验
	var req ArticleReq
	err := ctx.Bind(&req)
	if err != nil {
		u.l.Warn("参数有误")
		ctx.String(http.StatusBadRequest, "参数有误")
		return
	}
	// 取jwt claims
	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		u.l.Error("handler中未拿到 user claims")
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	// 业务处理
	id, err := u.svc.Store(ctx, req.reqToDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "OK",
			Data: id,
		})

	}
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "OK",
		Data: id,
	})
}

// Publish 发表文章
func (u *ArticleHandler) Publish(ctx *gin.Context) {
	// 参数接收 & 校验
	var req ArticleReq
	err := ctx.Bind(&req)
	if err != nil {
		u.l.Warn("参数有误")
		ctx.String(http.StatusBadRequest, "参数有误")
		return
	}
	// 取jwt claims
	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		u.l.Error("handler中未拿到 user claims")
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	// 业务处理
	id, err := u.svc.Publish(ctx, req.reqToDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 打日志
		//u.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "OK",
		Data: id,
	})
}

func (u *ArticleHandler) Withdraw(ctx *gin.Context) {
	// 参数接收 & 校验
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		u.l.Warn("参数有误")
		ctx.String(http.StatusBadRequest, "参数有误")
		return
	}
	// 取jwt claims
	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		u.l.Error("handler中未拿到 user claims")
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	// 业务处理
	err = u.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
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

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (req ArticleReq) reqToDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}
