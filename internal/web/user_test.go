package web

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webookpro/internal/domain"
	"webookpro/internal/service"
	svcmocks "webookpro/internal/service/mock"
)

// TestUserHandler_SignUp
func TestUserHandler_SignUp(t *testing.T) {
	testcases := []struct {
		Name     string
		reqBody  string
		mock     func(controller *gomock.Controller) service.UserService
		wantCode int
		wantBody string
	}{
		{
			Name: "注册成功",
			reqBody: `
				{
					"email":"123@qq.com",
					"password":"$pl3nd1D",
					"confirmPassword":"$pl3nd1D"
				}
            `,
			mock: func(controller *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(controller)
				userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(nil)
				return userSvc
			},
			wantBody: "注册成功",
			wantCode: http.StatusOK,
		},
		{
			Name: "参数绑定错误",
			reqBody: `
				{
					"email":"123@qq.com",
					"password":"$pl3nd1D",
				}
            `,
			mock: func(controller *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(controller)
				return userSvc
			},
			wantBody: "参数错误",
			wantCode: http.StatusBadRequest,
		},
		{
			Name: "邮箱格式错误",
			reqBody: `
				{
					"email":"123.com",
					"password":"$pl3nd1D",
					"confirmPassword":"$pl3nd1D"
				}
            `,
			mock: func(controller *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(controller)
				return userSvc
			},
			wantBody: "邮箱格式错误",
			wantCode: http.StatusOK,
		},
		{
			Name: "密码不一致",
			reqBody: `
				{
					"email":"123@qq.com",
					"password":"$pl3nd1D",
					"confirmPassword":"$pl3nd2D"
				}
            `,
			mock: func(controller *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(controller)
				return userSvc
			},
			wantBody: "密码不一致",
			wantCode: http.StatusOK,
		},
		{
			Name: "密码格式错误",
			reqBody: `
				{
					"email":"123@qq.com",
					"password":"000",
					"confirmPassword":"000"
				}
            `,
			mock: func(controller *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(controller)
				return userSvc
			},
			wantBody: "密码必须大于8位，包含数字、特殊字符",
			wantCode: http.StatusOK,
		},
		{
			Name: "邮箱重复",
			reqBody: `
				{
					"email":"123@qq.com",
					"password":"$pl3nd1D",
					"confirmPassword":"$pl3nd1D"
				}
            `,
			mock: func(controller *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(controller)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "$pl3nd1D",
				}).Return(service.ErrUserDuplicateEmail)
				return userSvc
			},
			wantBody: "邮箱重复，请换一个邮箱",
			wantCode: http.StatusOK,
		},
		{
			Name: "系统错误",
			reqBody: `
				{
					"email":"123@qq.com",
					"password":"$pl3nd1D",
					"confirmPassword":"$pl3nd1D"
				}
            `,
			mock: func(controller *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(controller)
				userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(errors.New("internal err"))
				return userSvc
			},
			wantBody: "系统错误",
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			server := gin.Default()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userSvc := tc.mock(ctrl)
			userHdl := NewUserHandler(userSvc, nil, nil)
			userHdl.RegisterRoutes(server)

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}
