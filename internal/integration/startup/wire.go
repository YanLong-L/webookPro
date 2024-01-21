//go:build wireinject

package startup

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

var thirdProvider = wire.NewSet(InitDB, InitRDB, ioc.InitLogger)

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

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(thirdProvider,
		dao.NewGORMArticleDAO,
		service.NewArticleService,
		web.NewArticleHandler,
		repository.NewCachedArticleRepository,
	)
	return &web.ArticleHandler{}
}
