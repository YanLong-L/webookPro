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
		dao.NewUserDAO,
		// cache 层
		cache.NewCodeCache, cache.NewRedisUserCache,
		// repo层
		repository.NewUserRepository, repository.NewCodeRepository,
		// service 层
		service.NewUserService, service.NewCodeService, memory.NewService,
		// handlers
		web.NewUserHandler,
		// middlewares
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return new(gin.Engine)
}
