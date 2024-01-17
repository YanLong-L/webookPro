package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"webookpro/internal/domain"
)

type jwtHandler struct {
}

// setJwtToken 设置jwt token
func (h *jwtHandler) setJwtToken(ctx *gin.Context, user domain.User) error {
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		Uid:       user.Id,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)), // 暂时设置成1分钟过期
		},
	})
	tokenStr, err := tokenObj.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	ctx.Header("x-jwt-token", tokenStr)
	return err
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
