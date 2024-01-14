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
	"webookpro/internal/service/sms/memory"
	"webookpro/internal/web"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRDB,
		// dao 层
		dao.NewGormUserDAO,
		// cache 层
		cache.NewRedisCodeCache, cache.NewRedisUserCache,
		// repo层
		repository.NewCachedUserRepository, repository.NewCachedCodeRepository,
		// service 层
		service.NewUserService, service.NewSMSCodeService, memory.NewService,
		// handlers
		web.NewUserHandler,
		// middlewares
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return new(gin.Engine)
}
