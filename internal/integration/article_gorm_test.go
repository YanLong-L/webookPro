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
	"webookpro/internal/domain"
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
	artHdl := startup.InitArticleHandler(article.NewGORMArticleDAO(s.db))
	// 注册好了路由
	artHdl.RegisterRoutes(s.server)
}

// TearDownTest 每一个都会执行
func (s *ArticleTestSuite) TearDownTest() {
	// 清空所有数据，并且自增主键恢复到 1
	s.db.Exec("TRUNCATE TABLE articles")
	s.db.Exec("TRUNCATE TABLE `published_articles`")

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
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
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
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
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

func (s *ArticleTestSuite) TestArticle_Publish() {
	t := s.T()

	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)
		req   Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
				// 什么也不需要做
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art article.Article
				s.db.Where("author_id = ?", 123).First(&art)
				assert.Equal(t, "hello，你好", art.Title)
				assert.Equal(t, "随便试试", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				var publishedArt article.PublishedArticle
				s.db.Where("author_id = ?", 123).First(&publishedArt)
				assert.Equal(t, "hello，你好", publishedArt.Title)
				assert.Equal(t, "随便试试", publishedArt.Content)
				assert.Equal(t, int64(123), publishedArt.AuthorId)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: Article{
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 2,
				Msg:  "OK",
				Data: 1,
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				// 模拟已经存在的帖子
				s.db.Create(&article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 123,
				})
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art article.Article
				s.db.Where("id = ?", 2).First(&art)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)
				var publishedArt article.PublishedArticle
				s.db.Where("id = ?", 2).First(&publishedArt)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 2,
				Code: 2,
				Msg:  "OK",
			},
		},
		{
			name: "更新帖子，并且重新发表",
			before: func(t *testing.T) {
				art := article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 123,
				}
				s.db.Create(&art)
				part := article.PublishedArticle{
					Article: art,
				}
				s.db.Create(&part)
			},
			after: func(t *testing.T) {
				var art article.Article
				s.db.Where("id = ?", 3).First(&art)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)

				var part article.PublishedArticle
				s.db.Where("id = ?", 3).First(&part)
				assert.Equal(t, "新的标题", part.Title)
				assert.Equal(t, "新的内容", part.Content)
				assert.Equal(t, int64(123), part.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				assert.True(t, part.Utime > 234)
			},
			req: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 2,
				Msg:  "OK",
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				art := article.Article{
					Id:      4,
					Title:   "我的标题",
					Content: "我的内容",
					Ctime:   456,
					Utime:   234,
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorId: 789,
				}
				s.db.Create(&art)
				part := article.PublishedArticle{
					Article: article.Article{
						Id:       4,
						Title:    "我的标题",
						Content:  "我的内容",
						Ctime:    456,
						Utime:    234,
						AuthorId: 789,
					},
				}
				s.db.Create(&part)
			},
			after: func(t *testing.T) {
				// 更新应该是失败了，数据没有发生变化
				var art article.Article
				s.db.Where("id = ?", 4).First(&art)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, int64(234), art.Utime)
				assert.Equal(t, int64(789), art.AuthorId)

				var part article.PublishedArticle
				// 数据没有变化
				s.db.Where("id = ?", 4).First(&part)
				assert.Equal(t, "我的标题", part.Title)
				assert.Equal(t, "我的内容", part.Content)
				assert.Equal(t, int64(789), part.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				assert.Equal(t, int64(234), part.Utime)
			},
			req: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// 不能有 error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			// 反序列化为结果
			// 利用泛型来限定结果必须是 int64
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
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
