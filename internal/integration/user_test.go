package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webookpro/internal/integration/startup"
	"webookpro/internal/ioc"
	"webookpro/internal/web"
)

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	// 既然是集成测试，那我需要一整个包含注册路由的 webserver，所以我可以直接从wire中拿
	server := startup.InitWebServer()
	rdb := ioc.InitRDB()
	testcases := []struct {
		name     string
		before   func() // 准备数据
		after    func() // 清理数据
		reqBody  string
		wantCode int
		wantResp web.Result
	}{
		{
			name: "验证码发送成功",
			reqBody: `
				{
					"phone":"13512345678"
				}
			`,
			wantCode: 200,
			wantResp: web.Result{
				Code: 2,
				Msg:  "发送成功",
			},
			after: func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				_, err := rdb.Del(ctx, "phone_code:login:13512345678").Result()
				_, err = rdb.Del(ctx, "phone_code:login:13512345678:cnt").Result()
				cancel()
				assert.NoError(t, err)
			},
			before: func() {

			},
		},
		{
			name: "参数绑定错误",
			reqBody: `
				{
					"phones":"13512345678
				}
			`,
			wantCode: 400,
			wantResp: web.Result{
				Code: 4,
				Msg:  "输入有误",
			},
			after: func() {

			},
			before: func() {

			},
		},
		{
			name: "手机号格式错误",
			reqBody: `
				{
					"phone":"13512345678910"
				}
			`,
			wantCode: 200,
			wantResp: web.Result{
				Code: 5,
				Msg:  "输入有误",
			},
			after: func() {

			},
			before: func() {

			},
		},
		{
			name: "发送太频繁",
			reqBody: `
				{
					"phone":"13512345678"
				}
			`,
			wantCode: 200,
			wantResp: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
			before: func() {
				// 在这里要准备数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				rdb.Set(ctx, "phone_code:login:13512345678", "123456", time.Minute*9+time.Second*30)
				rdb.Set(ctx, "phone_code:login:13512345678:cnt", 3, time.Minute*9+time.Second*30)
				cancel()
			},
			after: func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				_, err := rdb.Del(ctx, "phone_code:login:13512345678").Result()
				_, err = rdb.Del(ctx, "phone_code:login:13512345678:cnt").Result()
				cancel()
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			server.ServeHTTP(recorder, req)

			var res web.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, recorder.Code, tc.wantCode)
			assert.Equal(t, res, tc.wantResp)

			tc.after()
		})
	}
}
