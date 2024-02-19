package web

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
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
	svc     service.ArticleService
	intrSvc service.InteractiveService
	l       logger.Logger
	biz     string
}

func NewArticleHandler(svc service.ArticleService, intrSvc service.InteractiveService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		intrSvc: intrSvc,
		l:       l,
		biz:     "article",
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
	pub.POST("/like", ginx.WrapBodyAndToken[LikeReq,
		ijwt.UserClaims](u.Like))
}

// Like 点赞或取消点赞
func (u *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		err = u.intrSvc.Like(ctx, u.biz, req.Id, uc.Uid)
	} else {
		err = u.intrSvc.CancelLike(ctx, u.biz, req.Id, uc.Uid)
	}
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

// PubDetail 线上库帖子详情
func (u *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		u.l.Error("前端输入的 ID 不对", logger.Error(err))
		return
	}

	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)
	uc := ctx.MustGet("users").(ijwt.UserClaims)
	eg.Go(func() error {
		art, err = u.svc.GetPublishedById(ctx, id, uc.Uid)
		return err
	})
	eg.Go(func() error {
		// 在这里要获取文章的互动信息
		uc := ctx.MustGet("claims").(ijwt.UserClaims)
		// 这个地方可以容忍错误
		intr, err = u.intrSvc.Get(ctx, u.biz, id, uc.Uid)
		return err
	})
	// 在这儿等，要保证前面两个
	err = eg.Wait()
	if err != nil {
		// 代表查询出错了
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 加载完详情数据后，这里要阅读数 + 1
	go func() {
		// 开一个goroutine 异步去执行
		er := u.intrSvc.IncrReadCnt(ctx, u.biz, art.Id)
		if er != nil {
			u.l.Error("增加阅读计数失败",
				logger.Int64("artId", art.Id),
				logger.Error(er))
		}
	}()
	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:         art.Id,
			Title:      art.Title,
			Status:     art.Status.ToUint8(),
			Content:    art.Content,
			Author:     art.Author.Name, // 要把作者信息带出去
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			LikeCnt:    intr.LikeCnt,
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
		},
	})
}

// Detail 创作者帖子详情
func (u *ArticleHandler) Detail(ctx *gin.Context, claims ijwt.UserClaims) (ginx.Result, error) {
	paramId := ctx.Param("id")
	artId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: 4,
			Msg:  "系统错误",
		}, err
	}
	art, err := u.svc.GetById(ctx, artId)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	// 判定登录用户是否时这篇帖子的创作者
	if claims.Uid != art.Author.Id {
		return ginx.Result{
			Code: 4,
			// 也不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", claims.Uid)
	}
	return ginx.Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			//Author: art.Author
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	}, nil
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
