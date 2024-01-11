package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
	"webookpro/internal/web"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgorePath(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 校验jwt token
		header := ctx.GetHeader("Authorization")
		if header == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		headerSplit := strings.Split(header, " ")
		if len(headerSplit) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := headerSplit[1]
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid || uc.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 校验user agent
		if ctx.Request.UserAgent() != uc.UserAgent {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 此时token校验通过，判断token是否需要刷新，如需刷新则生成一个新的token
		if uc.ExpiresAt.Sub(time.Now()) < time.Second*50 { // 说明距离上次刷新已经过了10秒钟了，开始刷新
			// 刷新的token的本质就是生成一个新的token
			uc.ExpiresAt = jwt.NewNumericDate(uc.ExpiresAt.Add(time.Minute))
			newTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, uc)
			newTokenStr, _ := newTokenObj.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
			ctx.Header("x-jwt-token", newTokenStr)
		}
		ctx.Set("claims", uc)
	}
}
