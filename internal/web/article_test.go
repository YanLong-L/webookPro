package web

import (
	"bytes"
	"encoding/json"
	gin "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webookpro/internal/domain"
	"webookpro/internal/service"
	svcmocks "webookpro/internal/service/mock"
	"webookpro/internal/web/jwt"
	"webookpro/pkg/logger"
)

func TestArticleHandler_Publish(t *testing.T) {
	testcases := []struct {
		name     string
		mock     func(controller *gomock.Controller) service.ArticleService
		reqBody  string
		wantCode int
		wantRes  Result
	}{
		{
			name: "新建帖子，并发表成功",
			reqBody: `
				{"title":"我的标题","content":"我的内容"}
				`,
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				artSvc := svcmocks.NewMockArticleServcie(ctrl)
				artSvc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return artSvc
			},
			wantCode: 200,
			wantRes: Result{
				Code: 2,
				Msg:  "OK",
				Data: float64(1),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			server.Use(func(context *gin.Context) {
				context.Set("claims", &jwt.UserClaims{
					Uid: 123,
				})
			})

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			artSvc := tc.mock(ctrl)
			artHdl := NewArticleHandler(artSvc, &logger.NopLogger{})
			artHdl.RegisterRoutes(server)

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			server.ServeHTTP(recorder, req)

			var res Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
