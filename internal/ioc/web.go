package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
	"webookpro/internal/web"
	"webookpro/internal/web/middleware"
	"webookpro/pkg/ginx/middlewares/ratelimit"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()

	// 注册中间件
	server.Use(middlewares...)

	// 注册路由
	userHdl.RegisterRoutes(server)

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

func InitMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		rateLimitMiddleware(),
		corsMiddleware(),
		jwtMiddleware(),
	}

}

// corsMiddleware 跨域中间件
func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"x-jwt-token"},
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
func jwtMiddleware() gin.HandlerFunc {
	return middleware.NewLoginJWTMiddlewareBuilder().
		IgorePath("/users/login").
		IgorePath("/users/signup").
		IgorePath("/users/login_sms/code/send").
		IgorePath("/users/login_sms").
		Build()
}

// rateLimitMiddleware 限流中间件
func rateLimitMiddleware() gin.HandlerFunc {
	return ratelimit.NewBuilder(InitRDB(), time.Second, 100).Build()
}
