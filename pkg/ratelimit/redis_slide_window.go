package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed slide_window.lua
var luaSlideWindow string

type RedisSlideWindowLimiter struct {
	cmd      redis.Cmdable
	rate     int           // 阈值
	interval time.Duration // 窗口大小
	// interval 内允许 rate 个请求
	// 1s 内允许 3000 个请求
}

func NewRedisSlideWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlideWindowLimiter {
	return &RedisSlideWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}

}

func (r RedisSlideWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaSlideWindow, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}
