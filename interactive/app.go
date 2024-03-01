package main

import (
	"webookpro/pkg/ginx"
	"webookpro/pkg/grpcx"
	"webookpro/pkg/saramax"
)

type App struct {
	// 在这里，所有需要 main 函数控制启动、关闭的，都会在这里有一个
	// 核心就是为了控制生命周期
	server    *grpcx.Server
	consumers []saramax.Consumer
	webAdmin  *ginx.Server // 用来管理数据迁移的几个接口
}
