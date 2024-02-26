package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	// 初始化 viper配置中心
	InitViper()
	// 初始化logger
	InitLogger()
	//// 初始化Prometheus
	//InitPrometheus()
	//// 初始化opentelemetry
	//closeFunc := ioc.InitOTEL()
	// 初始化 app
	app := InitWebServer()

	// 开启所有消费者
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	// 开启所有定时任务
	app.cron.Start()

	// 开启web服务
	server := app.web
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}

	// 一分钟内shutdown trace provider
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	//closeFunc(ctx)

	ctx = app.cron.Stop()
	// 这边可以考虑超时强制退出，防止有些任务，执行特别长的时间
	tm := time.NewTimer(time.Minute * 10)
	select {
	case <-tm.C:
	case <-ctx.Done():
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

func InitPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()
}
