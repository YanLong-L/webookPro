package article

import (
	"context"
	"webookpro/internal/domain"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Upsert(ctx context.Context, article PublishedArticle) error
	SyncStatus(ctx context.Context, article Article, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, authorId int64, offset int, limit int) ([]Article, error)
}
