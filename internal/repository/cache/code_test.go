package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	redismocks "webookpro/internal/repository/cache/redismock"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testcases := []struct {
		name    string
		mock    func(controller *gomock.Controller) redis.Cmdable
		ctx     context.Context
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			name: "验证码设置成功",
			mock: func(controller *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(controller)
				mockRes := redis.NewCmdResult(int64(0), nil)
				cmd.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRes)
				return cmd
			},
			biz:     "login",
			phone:   "133",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "redis崩溃",
			mock: func(controller *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(controller)
				mockRes := redis.NewCmdResult(int64(0), errors.New("redis error"))
				cmd.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRes)
				return cmd
			},
			biz:     "login",
			phone:   "133",
			code:    "123456",
			wantErr: errors.New("redis error"),
		},
		{
			name: "发送太频繁",
			mock: func(controller *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(controller)
				mockRes := redis.NewCmdResult(int64(-1), nil)
				cmd.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRes)
				return cmd
			},
			biz:     "login",
			phone:   "133",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
		{
			name: "系统错误",
			mock: func(controller *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(controller)
				mockRes := redis.NewCmdResult(int64(-2), nil)
				cmd.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRes)
				return cmd
			},
			biz:     "login",
			phone:   "133",
			code:    "123456",
			wantErr: errors.New("系统错误"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctrl.Finish()
			redisCodeCache := NewRedisCodeCache(tc.mock(ctrl))
			err := redisCodeCache.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
