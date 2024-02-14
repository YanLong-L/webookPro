//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webookpro/internal/events/article"
	"webookpro/internal/ioc"
	"webookpro/internal/repository"
	article3 "webookpro/internal/repository/article"
	"webookpro/internal/repository/cache"
	"webookpro/internal/repository/dao"
	article2 "webookpro/internal/repository/dao/article"
	"webookpro/internal/service"
	"webookpro/internal/web"
	ijwt "webookpro/internal/web/jwt"
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRDB, ioc.InitLogger,
		ioc.InitKafka, ioc.NewConsumers, ioc.NewSyncProducer,
		// consumers
		article.NewInteractiveReadEventBatchConsumer,
		// producers
		article.NewKafkaProducer,
		// dao 层
		dao.NewGormUserDAO,
		dao.NewGORMInteractiveDAO,
		article2.NewGORMArticleDAO,
		// cache 层
		cache.NewRedisCodeCache,
		cache.NewRedisUserCache,
		cache.NewRedisInteractiveCache,
		cache.NewRedisArticleCache,
		// repo层
		repository.NewCachedUserRepository,
		repository.NewCachedCodeRepository,
		repository.NewCachedIntrRepository,
		article3.NewCachedArticleRepository,
		// service 层
		service.NewUserService,
		service.NewSMSCodeService,
		service.NewArticleService,
		service.NewInteractiveService,
		ioc.InitSMSService, ioc.InitWechatService,
		// handlers
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		web.NewArticleHandler,
		// middlewares
		ioc.InitMiddlewares,
		ioc.InitWebServer,
		ioc.InitLimiter,
		// 组装我这个结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
