//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	repository2 "webookpro/interactive/repository"
	cache2 "webookpro/interactive/repository/cache"
	dao2 "webookpro/interactive/repository/dao"
	service2 "webookpro/interactive/service"
	"webookpro/internal/ioc"
	"webookpro/internal/repository"
	"webookpro/internal/repository/article"
	"webookpro/internal/repository/cache"
	"webookpro/internal/repository/dao"
	article2 "webookpro/internal/repository/dao/article"
	"webookpro/internal/service"
	"webookpro/internal/web"
	ijwt "webookpro/internal/web/jwt"
)

var thirdProvider = wire.NewSet(InitDB, InitRDB, ioc.InitLogger)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedIntrRepository,
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		InitDB, InitRDB, ioc.InitLogger,
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

func InitArticleHandler(dao article2.ArticleDAO) *web.ArticleHandler {
	wire.Build(thirdProvider,
		service.NewArticleService,
		web.NewArticleHandler,
		article.NewCachedArticleRepository,
		cache.NewRedisArticleCache,
		interactiveSvcProvider,
	)
	return &web.ArticleHandler{}
}

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service2.NewInteractiveService(nil, nil)
}
