package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"webookpro/internal/integration/startup"
	"webookpro/internal/repository/dao/article"
	ijwt "webookpro/internal/web/jwt"
)

// ArticleTestSuite 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupSuite() {
	// 在所有测试执行之前，初始化一些内容
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
	})
	s.db = startup.InitDB()
	artHdl := startup.InitArticleHandler()
	// 注册好了路由
	artHdl.RegisterRoutes(s.server)
}

// TearDownTest 每一个都会执行
func (s *ArticleTestSuite) TearDownTest() {
	// 清空所有数据，并且自增主键恢复到 1
	s.db.Exec("TRUNCATE TABLE articles")
}

func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testcase := []struct {
		name string
		// 集成测试准备数据
		before func(t *testing.T)
		// 集成测试清理数据
		after    func(t *testing.T)
		req      Article
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子--保存成功",
			req: Article{
				Title:   "测试标题",
				Content: "测试内容",
			},
			wantCode: 200,
			wantRes: Result[int64]{
				Code: 2,
				Msg:  "OK",
				Data: 1,
			},
			before: func(t *testing.T) {
				s.db.Exec("TRUNCATE TABLE articles")
			},
			after: func(t *testing.T) {
				// 验证数据库
				var art article.Article
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, article.Article{
					Id:       1,
					Title:    "测试标题",
					Content:  "测试内容",
					AuthorId: 123,
				}, art)
				s.db.Exec("TRUNCATE TABLE articles")
			},
		},
		{
			name: "修改帖子--保存成功",
			req: Article{
				Id:      1,
				Title:   "测试标题",
				Content: "测试内容",
			},
			wantCode: 200,
			wantRes: Result[int64]{
				Code: 2,
				Msg:  "OK",
				Data: 1,
			},
			before: func(t *testing.T) {
				//s.db.Exec("TRUNCATE TABLE articles")
				// 我要先准备一条数据
				s.db.Create(&article.Article{
					Id:       1,
					Title:    "原标题",
					Content:  "原测试内容",
					AuthorId: 123,
					Ctime:    123,
					Utime:    123,
				})
			},
			after: func(t *testing.T) {
				// 验证数据库
				var art article.Article
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, art.Ctime, int64(123))
				// utime 要变但是 ctime不变
				assert.True(t, art.Utime > 123)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, article.Article{
					Id:       1,
					Title:    "测试标题",
					Content:  "测试内容",
					AuthorId: 123,
				}, art)
				s.db.Exec("TRUNCATE TABLE articles")
			},
		},
		{
			name: "试图修改别人的帖子",
			req: Article{
				Id:      1,
				Title:   "测试标题",
				Content: "测试内容",
			},
			wantCode: 200,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "OK",
				Data: 1,
			},
			before: func(t *testing.T) {
				//s.db.Exec("TRUNCATE TABLE articles")
				// 我要先准备一条数据
				s.db.Create(&article.Article{
					Id:       1,
					Title:    "原标题",
					Content:  "原测试内容",
					AuthorId: 456,
					Ctime:    123,
					Utime:    123,
				})
			},
			after: func(t *testing.T) {
				s.db.Exec("TRUNCATE TABLE articles")
			},
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			recorder := httptest.NewRecorder()
			reqbody, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer(reqbody))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			s.server.ServeHTTP(recorder, req)

			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, recorder.Code, tc.wantCode)
			assert.Equal(t, res, tc.wantRes)

			tc.after(t)
		})
	}
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
