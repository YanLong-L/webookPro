package startup

import (
	"github.com/redis/go-redis/v9"
	"webookpro/config"
)

// initRDB 初始化redis
func InitRDB() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
		//Addr: viper.GetString("redis.addr"),
	})
}
