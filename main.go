package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
	"webookpro/internal/repository"
	"webookpro/internal/repository/dao"
	"webookpro/internal/service"
	"webookpro/internal/web"
	"webookpro/internal/web/middleware"
	"webookpro/pkg/ginx/middlewares/ratelimit"
)

// initWebServer 初始化gin engine
func initWebServer() *gin.Engine {
	server := gin.Default()

	// 设置限流中间件
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	// 使用跨域中间件
	server.Use(cors.New(cors.Config{
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
	}))

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

	// 设置JWT中间件
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgorePath("/users/login").
		IgorePath("/users/signup").
		Build())

	return server
}

// initDB 初始化gormdb
func initDB() *gorm.DB {
	// 初始化db
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:13316)/webookpro"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// 初始化table
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

// initUser 初始化User服务
func initUser(db *gorm.DB, server *gin.Engine) {
	userDAO := dao.NewUserDAO(db)
	userRepo := repository.NewUserRepository(userDAO)
	userService := service.NewUserService(userRepo)
	userHandler := web.NewUserHandler(userService)
	userHandler.RegisterRoutes(server)
}

func main() {
	server := initWebServer()
	db := initDB()
	initUser(db, server)
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}
