package service

import (
	"context"
	"webookpro/internal/domain"
	"webookpro/internal/repository"
	"webookpro/pkg/logger"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	// Like 点赞
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	// CancelLike 取消点赞
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	// Collect 收藏
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.Logger
}

func NewInteractiveService(repo repository.InteractiveRepository, l logger.Logger) InteractiveService {
	return &interactiveService{
		repo: repo,
		l:    l,
	}
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func (i *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId, cid, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (i *interactiveService) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}
