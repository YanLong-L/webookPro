//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webookpro/interactive/events"
	"webookpro/interactive/grpc"
	"webookpro/interactive/ioc"
	"webookpro/interactive/repository"
	"webookpro/interactive/repository/cache"
	"webookpro/interactive/repository/dao"
	"webookpro/interactive/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitKafka,
	// 暂时不理会 consumer 怎么启动
	ioc.InitRedis,
	// 下面是数据迁移用到的
	ioc.InitDST,
	ioc.InitSRC,
	ioc.InitBizDB,
	ioc.InitDoubleWritePool,
	ioc.InitSyncProducer,
)

var migratorProvider = wire.NewSet(
	ioc.InitMigratorWeb,
	ioc.InitFixDataConsumer,
	ioc.InitMigradatorProducer)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedIntrRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitAPP() *App {
	wire.Build(interactiveSvcProvider,
		thirdPartySet,
		migratorProvider,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewConsumers,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
