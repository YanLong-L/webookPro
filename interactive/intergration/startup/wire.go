//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"webookpro/interactive/repository"
	cache2 "webookpro/interactive/repository/cache"
	dao2 "webookpro/interactive/repository/dao"
	service2 "webookpro/interactive/service"
)

var thirdProvider = wire.NewSet(InitRedis,
	InitTestDB, InitLog)
var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository.NewCachedIntrRepository,
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
)

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service2.NewInteractiveService(nil, nil)
}
