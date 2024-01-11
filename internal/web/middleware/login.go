package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgorePath(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		session := sessions.Default(ctx)
		session.Options(sessions.Options{
			MaxAge: 60,
		})
		userId := session.Get("user_id")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		// 取update_time 和 now比 查看是否过了30秒，如果过了30s就续约
		updateTime := session.Get("update_time")
		if updateTime == nil { // 说明还没有续约过
			session.Set("user_id", userId)
			session.Set("update_time", time.Now().UnixMilli())
			if err := session.Save(); err != nil {
				panic(err)
			}
			return
		}
		updateTimeVal, _ := updateTime.(int64)
		if time.Now().UnixMilli()-updateTimeVal > 30*1000 {
			// 说明过了一分钟了，直接续约
			session.Set("user_id", userId)
			session.Set("update_time", time.Now().UnixMilli())
			if err := session.Save(); err != nil {
				panic(err)
			}
		}
	}
}

// CheckLogin  初始版本的LoginMiddleware
func CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/users/login" ||
			ctx.Request.URL.Path == "users/signup" {
			return
		}
		session := sessions.Default(ctx)
		user_id := session.Get("user_id")
		if user_id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
