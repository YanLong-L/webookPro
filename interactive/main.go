package main

import (
	"github.com/spf13/viper"
	"log"
)

func main() {
	InitViper()
	app := InitAPP()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	go func() {
		err := app.webAdmin.Start()
		log.Println(err)
	}()
	err := app.server.Serve()
	log.Println(err)
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
