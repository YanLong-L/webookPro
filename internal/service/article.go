package service

import (
	"context"
	"webookpro/internal/domain"
	"webookpro/internal/repository"
)

type ArticleServcie interface {
	Store(ctx context.Context, article domain.Article) (int64, error)
}

type articleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) ArticleServcie {
	return &articleService{
		repo: repo,
	}

}

func (s *articleService) Store(ctx context.Context, article domain.Article) (int64, error) {
	id, err := s.repo.Create(ctx, article)
	return id, err
}
