package repository

import (
	"context"
	"webookpro/internal/domain"
	"webookpro/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) (domain.Article, error)
}

type CachedRankingRepository struct {
	// 使用具体实现，可读性更好，对测试不友好，因为没有面向接口编程
	redis *cache.RankingRedisCache
	local *cache.RankingLocalCache
}

func NewCachedRankingRepository(redis *cache.RankingRedisCache, local *cache.RankingLocalCache) *CachedRankingRepository {
	return &CachedRankingRepository{redis: redis, local: local}
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	// 先设置本地缓存，因为本地缓存很难出错
	_ = c.local.Set(ctx, arts)
	// 设置redis缓存
	return c.redis.Set(ctx, arts)
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	// 先查本地缓存
	data, err := c.local.Get(ctx)
	if err == nil {
		// 如果查到了就直接返回
		return data, nil
	}
	// 本地缓存没查到，查redis
	data, err = c.redis.Get(ctx)
	if err == nil {
		// redis 查到了，回写本地缓存
		_ = c.local.Set(ctx, data)

	} else {
		// redis 也没查到，读老数据
		return c.local.ForceGet(ctx)
	}
	return data, err
}
