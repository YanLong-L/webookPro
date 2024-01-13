//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(webookpro-live-mysql:11309)/webook",
	},
	Redis: RedisConfig{
		Addr: "webookpro-live-redis:11479",
	},
}
