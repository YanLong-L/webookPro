package repository

import (
	"context"
	"webookpro/internal/repository/cache"
)

var (
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
	ErrUnknownForCode         = cache.ErrUnknownForCode
)

type CodeRepository struct {
	cache *cache.CodeCache
}

func NewCodeRepository(cache *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: cache,
	}
}

// Store 缓存验证码
func (repo *CodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

// Verfiy 校验验证码
func (repo *CodeRepository) Verfiy(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCode)

}
