package web

import (
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"webookpro/internal/domain"
	"webookpro/internal/service"
	ijwt "webookpro/internal/web/jwt"
	"webookpro/pkg/ginx"
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
	ug.POST("/list", ginx.WrapBodyAndToken[ListReq, ijwt.UserClaims](u.List))
	ug.GET("/detail/:id", ginx.WrapToken[ijwt.UserClaims](u.Detail))

	pub := ug.Group("/pub")
	pub.GET("/:id", u.PubDetail)
}

func (u *ArticleHandler) Detail(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {

}

// List 创作者文章列表
func (u *ArticleHandler) List(ctx *gin.Context, req ListReq, claims ijwt.UserClaims) (ginx.Result, error) {
	var res []domain.Article
	res, err := u.svc.List(ctx, claims.Uid, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Data: res,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res,
			func(idx int, src domain.Article) ArticleVO {
				return ArticleVO{
					Id:       src.Id,
					Title:    src.Title,
					Abstract: src.Abstract(),
					Status:   src.Status.ToUint8(),
					// 这个列表请求，不需要返回内容
					//Content: src.Content,
					// 这个是创作者看自己的文章列表，也不需要这个字段
					//Author: src.Author
					Ctime: src.Ctime.Format(time.DateTime),
					Utime: src.Utime.Format(time.DateTime),
				}
			}),
	}, nil
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
	id, err := u.svc.Store(ctx, req.toDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
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
	id, err := u.svc.Publish(ctx, req.toDomain(claims.Uid))
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
