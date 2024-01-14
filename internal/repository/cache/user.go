package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webookpro/internal/domain"
)

var (
	ErrKeyNotExist = redis.Nil
)

type UserCache interface {
	Set(ctx context.Context, user domain.User) error
	Get(ctx context.Context, id int64) (domain.User, error)
	key(ctx context.Context, id int64) string
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (uc *RedisUserCache) Set(ctx context.Context, user domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return uc.cmd.Set(ctx, uc.key(ctx, user.Id), data, uc.expiration).Err()
}

func (uc *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	data, err := uc.cmd.Get(ctx, uc.key(ctx, id)).Result()
	if err != nil {
		return domain.User{}, err
	}
	var user domain.User
	err = json.Unmarshal([]byte(data), &user)
	return user, err
}

func (uc *RedisUserCache) key(ctx context.Context, id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
