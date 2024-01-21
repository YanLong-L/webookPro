package main

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	InitViper()
	InitLogger()
	server := InitWebServer()
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func InitViper() {
	// SetConfigName 配置文件的名字，但是不包括文件扩展名
	viper.SetConfigName("dev")
	// SetConfigType 告诉viper 我的配置文件用的是 yaml的格式
	viper.SetConfigType("yaml")
	// 当前工作目录下的 config 子目录
	viper.AddConfigPath("./config")
	// 读取配置到viper里面
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func InitLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
