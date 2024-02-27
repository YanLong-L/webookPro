//go:build wireinject

package interactive

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

var thirdPartySet = wire.NewSet(ioc.InitDB,
	ioc.InitLogger,
	ioc.InitKafka,
	// 暂时不理会 consumer 怎么启动
	ioc.InitRedis)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedIntrRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitAPP() *App {
	wire.Build(interactiveSvcProvider,
		thirdPartySet,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewConsumers,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
