//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webookpro/internal/ioc"
	"webookpro/internal/repository"
	"webookpro/internal/repository/cache"
	"webookpro/internal/repository/dao"
	"webookpro/internal/service"
	"webookpro/internal/web"
	ijwt "webookpro/internal/web/jwt"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRDB, ioc.InitLogger,
		// dao 层
		dao.NewGormUserDAO,
		// cache 层
		cache.NewRedisCodeCache, cache.NewRedisUserCache,
		// repo层
		repository.NewCachedUserRepository, repository.NewCachedCodeRepository,
		// service 层
		service.NewUserService, service.NewSMSCodeService,
		ioc.InitSMSService, ioc.InitWechatService,
		// handlers
		web.NewUserHandler, web.NewOAuth2WechatHandler, ijwt.NewRedisJWTHandler,
		// middlewares
		ioc.InitMiddlewares,
		ioc.InitWebServer,
		ioc.InitLimiter,
	)
	return new(gin.Engine)
}
