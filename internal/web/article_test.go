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
	"webookpro/internal/service"
	"webookpro/pkg/logger"
)

func TestArticleHandler_Publish(t *testing.T) {
	testcases := []struct {
		name     string
		mock     func(controller *gomock.Controller) service.ArticleServcie
		reqBody  string
		wantCode int
		wantRes  Result
	}{
		{
			name: "发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleServcie {

			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()

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
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
