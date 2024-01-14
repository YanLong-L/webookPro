package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooMany        = errors.New("发送验证码太频繁")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
	ErrUnknownForCode         = errors.New("发送验证码未知错误")
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
	Key(ctx context.Context, biz string, phone string) string
}

type RedisCodeCache struct {
	cmd redis.Cmdable
}

//go:embed lua/set_code.lua
var setCodeLua string

//go:embed lua/verify_code.lua
var verifyCodeLua string

func NewRedisCodeCache(cmd redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		cmd: cmd,
	}
}

// Set 将验证码记入缓存
func (cache *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := cache.cmd.Eval(ctx, setCodeLua, []string{cache.Key(ctx, biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0: // 设置成功，符合预期
		return nil
	case -1: // 发送太频繁
		return ErrCodeSendTooMany
	default:
		return errors.New("系统错误")
	}
}

// Verify 校验验证码
func (cache *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := cache.cmd.Eval(ctx, verifyCodeLua, []string{cache.Key(ctx, biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
	}
	return false, ErrUnknownForCode
}

// Key 生成验证码的redis key
func (svc *RedisCodeCache) Key(ctx context.Context, biz string, phone string) string {
	return fmt.Sprintf("phone_code%s:%s", biz, phone)
}
