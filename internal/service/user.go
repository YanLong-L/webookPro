package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"webookpro/internal/domain"
	"webookpro/internal/repository"
)

var (
	ErrUserDuplicateEmail    error = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword error = errors.New("用户名或密码不对")
)

type UserService interface {
	Login(ctx context.Context, email, password string) (domain.User, error)
	SignUp(ctx context.Context, user domain.User) error
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreate(ctx *gin.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WeChatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// Login
func (u *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 查找用户是否存在
	user, err := u.repo.FindByEmail(ctx, email)
	if err == gorm.ErrRecordNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 校验用户名密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return user, err
}

// SignUp
func (u *userService) SignUp(ctx context.Context, user domain.User) error {
	// 密码加密
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(bcryptPassword)
	return u.repo.Create(ctx, user)
}

// Profile
func (u *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	user, err := u.repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err

	}
	return user, err
}

// FindOrCreate 通过phone查找用户，找不到就新建一个
func (u *userService) FindOrCreate(ctx *gin.Context, phone string) (domain.User, error) {
	user, err := u.repo.FindByPhone(ctx, phone)
	if err == nil {
		// 说明查找到了，直接返回
		return user, nil
	}
	// 没找到，新建一个并返回
	err = u.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	if err != nil {
		return domain.User{}, err
	}
	return u.repo.FindByPhone(ctx, phone)
}

// FindOrCreateByWechat 通过openId查找用户并返回用户信息，没有就创建一个再返回用户信息
func (u *userService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WeChatInfo) (domain.User, error) {
	user, err := u.repo.FindByWechat(ctx, wechatInfo.OpenID)
	if err == nil {
		// 说明查找到了，直接返回
		return user, nil
	}
	// 没找到，新建一个并返回
	err = u.repo.Create(ctx, domain.User{
		Wechat: wechatInfo,
	})
	if err != nil {
		return domain.User{}, err
	}
	return u.repo.FindByWechat(ctx, wechatInfo.OpenID)
}
