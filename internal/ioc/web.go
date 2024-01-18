package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webookpro/internal/web"
	ijwt "webookpro/internal/web/jwt"
	"webookpro/internal/web/middleware"
	"webookpro/pkg/ginx/middlewares/ratelimit"
	limit "webookpro/pkg/ratelimit"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHdl *web.UserHandler, wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()

	// 注册中间件
	server.Use(middlewares...)

	// 注册路由
	userHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)

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

func InitMiddlewares(limiter limit.Limiter, jwtHdl ijwt.JwtHandler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		rateLimitMiddleware(limiter),
		corsMiddleware(),
		jwtMiddleware(jwtHdl),
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
		Build()
}

// rateLimitMiddleware 限流中间件
func rateLimitMiddleware(limiter limit.Limiter) gin.HandlerFunc {
	return ratelimit.NewBuilder(InitRDB(), limiter).Build()
}

func InitLimiter(cmd redis.Cmdable) limit.Limiter {
	return limit.NewRedisSlideWindowLimiter(cmd, time.Second, 100)
}
