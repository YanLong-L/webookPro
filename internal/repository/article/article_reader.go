package article

import (
	"context"
	"webookpro/internal/domain"
)

type ArticleReaderRepository interface {
	// Save 有就更新，没有就新建，即 upsert 的语义
	Store(ctx context.Context, article domain.Article) (int64, error)
}
