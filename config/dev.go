//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:mysql@tcp(127.0.0.1:3307)/webookpro",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
