// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package integration

import (
	"github.com/gin-gonic/gin"
	"webookpro/internal/ioc"
	"webookpro/internal/repository"
	"webookpro/internal/repository/cache"
	"webookpro/internal/repository/dao"
	"webookpro/internal/service"
	"webookpro/internal/service/sms/memory"
	"webookpro/internal/web"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	v := ioc.InitMiddlewares()
	db := ioc.InitDB()
	userDAO := dao.NewGormUserDAO(db)
	cmdable := ioc.InitRDB()
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	smsService := memory.NewService()
	codeService := service.NewSMSCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService)
	engine := ioc.InitWebServer(v, userHandler)
	return engine
}
