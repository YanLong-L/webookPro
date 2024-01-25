package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"testing"
	"webookpro/internal/domain"
	"webookpro/internal/repository"
	repomocks "webookpro/internal/repository/mock"
	"webookpro/pkg/logger"
)

func TestUserService_Login(t *testing.T) {
	testcases := []struct {
		name     string
		mock     func(controller *gomock.Controller) repository.UserRepository
		email    string
		password string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(controller *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(controller)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Password: "$2a$10$RV9/taBHXln82KX6hlS9ie1JXK1va9vSKItLzrVVPO5dlRCNOEPMG",
				}, nil)
				return userRepo
			},
			email:    "123@qq.com",
			password: "$pl3nd1D",
			wantUser: domain.User{
				Password: "$2a$10$RV9/taBHXln82KX6hlS9ie1JXK1va9vSKItLzrVVPO5dlRCNOEPMG",
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(controller *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(controller)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, gorm.ErrRecordNotFound)
				return userRepo
			},
			email:    "123@qq.com",
			password: "$pl3nd1D",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "密码不正确",
			mock: func(controller *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(controller)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, ErrInvalidUserOrPassword)
				return userRepo
			},
			email:    "123@qq.com",
			password: "xxx",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(controller *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(controller)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, errors.New("数据库错误"))
				return userRepo
			},
			email:    "123@qq.com",
			password: "$pl3nd1D",
			wantUser: domain.User{},
			wantErr:  errors.New("数据库错误"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userRepo := tc.mock(ctrl)
			userSvc := NewUserService(userRepo, &logger.NopLogger{})
			user, err := userSvc.Login(context.Background(), tc.email, tc.password)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_Encrypt(t *testing.T) {
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte("$pl3nd1D"), bcrypt.DefaultCost)
	require.NoError(t, err)
	fmt.Println(string(bcryptPassword)) // $2a$10$RV9/taBHXln82KX6hlS9ie1JXK1va9vSKItLzrVVPO5dlRCNOEPMG
}
