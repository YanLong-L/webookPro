package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"webookpro/internal/service/sms"
)

type AuthSMSService struct {
	sms sms.Service
	key string
}

func NewAuthSMSService(sms sms.Service, key string) *AuthSMSService {
	return &AuthSMSService{
		sms: sms,
		key: key,
	}

}

// 这时候 biz就代表 发送方带来的token
func (a AuthSMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	// 根据传过来的key,验证token是否正确，如何token is valid，就可以发送
	var claims Claims
	token, err := jwt.ParseWithClaims(biz, &claims, func(token *jwt.Token) (interface{}, error) {
		return a.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token无效")
	}
	err = a.sms.Send(ctx, claims.Tpl, args, numbers...)
	return err
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}
