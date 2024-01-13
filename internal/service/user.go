package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"webookpro/internal/domain"
	"webookpro/internal/repository"
)

var (
	ErrUserDuplicateEmail    error = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword error = errors.New("用户名或密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// Login
func (u *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
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
func (u *UserService) SignUp(ctx context.Context, user domain.User) error {
	// 密码加密
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(bcryptPassword)
	return u.repo.Create(ctx, user)
}

// Profile
func (u *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	user, err := u.repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err

	}
	return user, err
}
