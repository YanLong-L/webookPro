package repository

import (
	"context"
	"webookpro/internal/domain"
	"webookpro/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
}

type CachedArticleRepository struct {
	dao dao.ArticleDAO
}

func NewCachedArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func (r *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return r.dao.Insert(ctx, dao.Article{
		AuthorId: article.Author.Id,
		Content:  article.Content,
		Title:    article.Title,
	})
}
