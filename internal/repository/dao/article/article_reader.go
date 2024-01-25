package article

import (
	"context"
	"gorm.io/gorm"
)

type ArticleReaderDAO interface {
	Upsert(ctx context.Context, article PublishedArticle) error
}

func NewArticleReaderDAO(db *gorm.DB) ArticleReaderDAO {
	panic("")
}
