package ioc

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"strings"
	"time"
	"webookpro/internal/web"
	ijwt "webookpro/internal/web/jwt"
	"webookpro/internal/web/middleware"
	"webookpro/pkg/ginx"
	glogger "webookpro/pkg/ginx/middlewares/logger"
	"webookpro/pkg/ginx/middlewares/metric"
	"webookpro/pkg/ginx/middlewares/ratelimit"
	"webookpro/pkg/logger"
	limit "webookpro/pkg/ratelimit"
)

func InitWebServer(middlewares []gin.HandlerFunc,
	userHdl *web.UserHandler,
	wechatHdl *web.OAuth2WechatHandler,
	articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()

	// 注册中间件
	server.Use(middlewares...)

	// 注册路由
	userHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)

	// 设置session
	//store := cookie.NewStore([]byte("secret"))
	//store, _ := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	//server.Use(sessions.Sessions("wb_ssid", store))
	//// 中间件校验session
	////server.Use(middleware.CheckLogin())
	//server.Use(middleware.NewLoginMiddlewareBuilder().
	//	IgorePath("/users/login").
	//	IgorePath("/users/signup").
	//	Build())

	return server
}

func InitMiddlewares(limiter limit.Limiter, jwtHdl ijwt.JwtHandler, l logger.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		rateLimitMiddleware(limiter),
		//logMiddleware(l),
		corsMiddleware(),
		jwtMiddleware(jwtHdl),
		metricsMiddleware(),
		otelgin.Middleware("webookpro"),
	}

}

// corsMiddleware 跨域中间件
func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"x-jwt-token", "x-fresh-token"},
		// 是否允许你带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 你的开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	})
}

// jwtMiddleware JWT 校验中间件
func jwtMiddleware(jwtHdl ijwt.JwtHandler) gin.HandlerFunc {
	return middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
		IgorePath("/users/login").
		IgorePath("/users/signup").
		IgorePath("/users/login_sms/code/send").
		IgorePath("/users/login_sms").
		IgorePath("/users/refresh_token").
		IgorePath("/oauth2/wechat/authurl").
		IgorePath("/oauth2/wechat/callback").
		IgorePath("/articles/edit").
		Build()
}

// rateLimitMiddleware 限流中间件
func rateLimitMiddleware(limiter limit.Limiter) gin.HandlerFunc {
	return ratelimit.NewBuilder(InitRDB(), limiter).Build()
}

// prometheus web 中间件
func metricsMiddleware() gin.HandlerFunc {
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return (&metric.MiddlewareBuilder{
		Namespace:  "geekbang",
		Subsystem:  "webookpro",
		Name:       "gin_http",
		Help:       "统计 GIN 的 HTTP 接口",
		InstanceID: "my-instance-1",
	}).Build()
}

func logMiddleware(l logger.Logger) gin.HandlerFunc {
	return glogger.NewBuilder(func(ctx context.Context, al *glogger.AccessLog) {
		l.Debug("Gin Http", logger.Field{Key: "AccessLog", Value: al})
	}).AllowReqBody(true).AllowRespBody().Build()
}

func InitLimiter(cmd redis.Cmdable) limit.Limiter {
	return limit.NewRedisSlideWindowLimiter(cmd, time.Second, 100)
}
