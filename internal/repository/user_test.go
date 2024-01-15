package repository

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"webookpro/internal/domain"
	"webookpro/internal/repository/cache"
	cachemocks "webookpro/internal/repository/cache/mock"
	"webookpro/internal/repository/dao"
	daomocks "webookpro/internal/repository/dao/mock"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	// 111ms.11111ns
	now := time.Now()
	// 你要去掉毫秒以外的部分
	// 111ms
	now = time.UnixMilli(now.UnixMilli())
	testcases := []struct {
		name     string
		id       int64
		mock     func(controller *gomock.Controller) (dao.UserDAO, cache.UserCache)
		wantUser domain.User
		wantErr  error
		ctx      context.Context
	}{
		{
			name: "缓存命中成功",
			id:   1,
			mock: func(controller *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uc := cachemocks.NewMockUserCache(controller)
				ud := daomocks.NewMockUserDAO(controller)
				uc.EXPECT().Get(gomock.Any(), int64(1)).Return(domain.User{
					Id: 1,
				}, nil)
				return ud, uc
			},
			wantErr: nil,
			wantUser: domain.User{
				Id: 1,
			},
			ctx: context.Background(),
		},
		{
			name: "缓存未命中，走数据库查询成功",
			id:   1,
			mock: func(controller *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uc := cachemocks.NewMockUserCache(controller)
				ud := daomocks.NewMockUserDAO(controller)
				uc.EXPECT().Get(gomock.Any(), int64(1)).Return(domain.User{}, ErrKeyNotExist)
				ud.EXPECT().FindById(gomock.Any(), int64(1)).Return(dao.User{
					Id:    1,
					Ctime: now.UnixMilli(),
					Utime: now.UnixMilli(),
				}, nil)
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:    1,
					Ctime: now,
				}).Return(nil)

				return ud, uc
			},
			wantErr: nil,
			wantUser: domain.User{
				Id:    1,
				Ctime: now,
			},
			ctx: context.Background(),
		},
		{
			name: "缓存未命中，走数据库查询也失败",
			id:   1,
			mock: func(controller *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uc := cachemocks.NewMockUserCache(controller)
				ud := daomocks.NewMockUserDAO(controller)
				uc.EXPECT().Get(gomock.Any(), int64(1)).Return(domain.User{}, errors.New("redis崩溃"))

				return ud, uc
			},
			wantErr:  errors.New("redis崩溃"),
			wantUser: domain.User{},
			ctx:      context.Background(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ud, uc := tc.mock(ctrl)
			userRepo := NewCachedUserRepository(ud, uc)
			user, err := userRepo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
