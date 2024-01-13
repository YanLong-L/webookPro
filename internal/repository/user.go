package repository

import (
	"context"
	"webookpro/internal/domain"
	"webookpro/internal/repository/cache"
	"webookpro/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail error = dao.ErrUserDuplicateEmail
	ErrKeyNotExist        error = cache.ErrKeyNotExist
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.RedisUserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.RedisUserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (u *UserRepository) Create(ctx context.Context, user domain.User) error {
	return u.dao.Insert(ctx, dao.User{
		Email:    user.Email,
		Password: user.Password,
	})
}

func (u *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
	}, err
}

func (u *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先查缓存
	user, err := u.cache.Get(ctx, id)
	switch err {
	case nil:
		//代表一定有数据
		return user, err
	case ErrKeyNotExist:
		// 代表key 确实不存在，查询数据库并写入缓存
		ue, err := u.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}
		// return前写入缓存
		user = domain.User{
			Id:       ue.Id,
			Email:    ue.Email,
			Password: ue.Password,
		}
		var _ = u.cache.Set(ctx, user)
		return user, err

	default:
		// 此时说明redis 异常，为保护数据库，不继续查询数据库
		return domain.User{}, err
	}

}
