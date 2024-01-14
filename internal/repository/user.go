package repository

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"time"
	"webookpro/internal/domain"
	"webookpro/internal/repository/cache"
	"webookpro/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail error = dao.ErrUserDuplicateEmail
	ErrKeyNotExist        error = cache.ErrKeyNotExist
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByPhone(ctx *gin.Context, phone string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCachedUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (u *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return u.dao.Insert(ctx, u.domainToEntity(user))
}

func (u *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return u.entityToDomain(user), err
}

func (u *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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
		user = u.entityToDomain(ue)
		var _ = u.cache.Set(ctx, user)
		return user, err

	default:
		// 此时说明redis 异常，为保护数据库，不继续查询数据库
		return domain.User{}, err
	}

}

func (u *CachedUserRepository) FindByPhone(ctx *gin.Context, phone string) (domain.User, error) {
	user, err := u.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return u.entityToDomain(user), err
}

// entityToDomain 实体对象转领域对象
func (u *CachedUserRepository) entityToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Password: user.Password,
		Ctime:    time.UnixMilli(user.Ctime),
		Phone:    user.Phone.String,
	}
}

// domainToEntity 领域对象转实体对象
func (u *CachedUserRepository) domainToEntity(user domain.User) dao.User {
	return dao.User{
		Id: user.Id,
		Email: sql.NullString{
			String: user.Email,
			// 我确实有手机号
			Valid: user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password: user.Password,
		Ctime:    user.Ctime.UnixMilli(),
	}
}
