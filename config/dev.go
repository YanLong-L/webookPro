//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(127.0.0.1:13316)/webookpro",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
