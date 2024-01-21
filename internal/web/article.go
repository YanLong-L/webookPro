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
}

// Edit 创作者编辑一篇文章并保存
func (u *ArticleHandler) Edit(ctx *gin.Context) {
	// 参数接收 & 校验
	type Req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
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
	id, err := u.svc.Store(ctx, domain.Article{
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
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
