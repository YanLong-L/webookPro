package repository

import (
	"context"
	"webookpro/internal/domain"
	"webookpro/internal/repository/cache"
	"webookpro/internal/repository/dao"
	"webookpro/pkg/logger"
)

//go:generate mockgen -source=./interactive.go -package=repomocks -destination=mocks/interactive.mock.go InteractiveRepository

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
}

type CachedIntrRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.Logger
}

func NewCachedIntrRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache, l logger.Logger) InteractiveRepository {
	return &CachedIntrRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (c *CachedIntrRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CachedIntrRepository) IncrLike(ctx context.Context, biz string, bizId, uid int64) error {
	// 先插入点赞 然后更新点赞计数  然后更新缓存
	err := c.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *CachedIntrRepository) DecrLike(ctx context.Context, biz string, bizId, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *CachedIntrRepository) AddCollectionItem(ctx context.Context, biz string, bizId, cid int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (c *CachedIntrRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CachedIntrRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	panic("implement me ")
}

func (c *CachedIntrRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	//TODO implement me
	panic("implement me")
}
