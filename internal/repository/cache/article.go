package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"webookpro/internal/domain"
)

type ArticleCache interface {
	// GetFirstPage 只缓存第第一页的数据
	// 并且不缓存整个 Content
	GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error
	DelFirstPage(ctx context.Context, author int64) error

	Set(ctx context.Context, art domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)

	// SetPub 正常来说，创作者和读者的 Redis 集群要分开，因为读者是一个核心中的核心
	SetPub(ctx context.Context, article domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}

type RedisArticleCache struct {
	client *redis.Cmdable
}

func (r RedisArticleCache) GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error) {
	//TODO implement me
	panic("implement me")
}

func (r RedisArticleCache) SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisArticleCache) DelFirstPage(ctx context.Context, author int64) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisArticleCache) Set(ctx context.Context, art domain.Article) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	//TODO implement me
	panic("implement me")
}

func (r RedisArticleCache) SetPub(ctx context.Context, article domain.Article) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	//TODO implement me
	panic("implement me")
}

// 创作端的缓存设置
func (r *RedisArticleCache) authorArtKey(id int64) string {
	return fmt.Sprintf("article:author:%d", id)
}

// 读者端的缓存设置
func (r *RedisArticleCache) readerArtKey(id int64) string {
	return fmt.Sprintf("article:reader:%d", id)
}

func (r *RedisArticleCache) firstPageKey(author int64) string {
	return fmt.Sprintf("article:first_page:%d", author)
}
