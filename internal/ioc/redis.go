package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// InitRDB 初始化redis
func InitRDB() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		//Addr: config.Config.Redis.Addr,
		Addr: viper.GetString("redis.addr"),
	})
}

// InitRLockClient 初始化Redis 分布式锁
func InitRLockClient(cmd redis.Cmdable) *rlock.Client {
	return rlock.NewClient(cmd)
}
