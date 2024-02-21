package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
	"webookpro/internal/domain"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingRedisCache struct {
	client redis.Cmdable
	key    string
}

func NewRedisRankingCache(client redis.Cmdable) *RankingRedisCache {
	return &RankingRedisCache{client: client, key: "ranking"}
}

func (r RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	// 可以预期的是， 当榜单缓存之后，用户极有可能会访问榜单中的文章
	// 因此这里可以做一个业务预加载，即在缓存榜单时，顺便将榜单中的文章内容也缓存起来

	// 因为作为一个榜单，是不需要缓存文章内容的，为了节约内存，这里把文章内容置为空
	for i := 0; i < len(arts); i++ {
		arts[i].Content = ""
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	// 注意这里的过期时间最好是超过计算热榜的时间（包括重试在内）
	// 甚至也可以永不过期，反正计算后会覆盖
	return r.client.Set(ctx, r.key, val, time.Minute*10).Err()
}

func (r RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	data, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(data, &res)
	return res, err
}
