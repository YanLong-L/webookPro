package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webookpro/internal/domain"
	"webookpro/internal/repository/article"
	artrepomocks "webookpro/internal/repository/article/mocks"
	"webookpro/pkg/logger"
)

func TestArticleService_PublishV1(t *testing.T) {
	testcases := []struct {
		name       string
		article    domain.Article
		authorRepo func(ctrl *gomock.Controller) article.ArticleAuthorRepository
		readerRepo func(ctrl *gomock.Controller) article.ArticleReaderRepository
		wantId     int64
		wantErr    error
	}{
		{
			name: "新建帖子，并发表成功",
			article: domain.Article{
				Title:   "测试标题",
				Content: "测试内容",
			},
			authorRepo: func(ctrl *gomock.Controller) article.ArticleAuthorRepository {
				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(int64(1), nil)
				return authorRepo
			},
			readerRepo: func(ctrl *gomock.Controller) article.ArticleReaderRepository {
				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Store(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(int64(1), nil)
				return readerRepo
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "更新帖子，并发表成功",
			article: domain.Article{
				Id:      1,
				Title:   "测试标题",
				Content: "测试内容",
			},
			authorRepo: func(ctrl *gomock.Controller) article.ArticleAuthorRepository {
				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(nil)
				return authorRepo
			},
			readerRepo: func(ctrl *gomock.Controller) article.ArticleReaderRepository {
				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Store(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(int64(1), nil)
				return readerRepo
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "新建帖子，保存到制作库失败",
			article: domain.Article{
				Title:   "测试标题",
				Content: "测试内容",
			},
			authorRepo: func(ctrl *gomock.Controller) article.ArticleAuthorRepository {
				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(int64(0), errors.New("mock error"))
				return authorRepo
			},
			readerRepo: func(ctrl *gomock.Controller) article.ArticleReaderRepository {
				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
				return readerRepo
			},
			wantId:  0,
			wantErr: errors.New("mock error"),
		},
		{
			name: "更新帖子，保存到制作库失败",
			article: domain.Article{
				Id:      1,
				Title:   "测试标题",
				Content: "测试内容",
			},
			authorRepo: func(ctrl *gomock.Controller) article.ArticleAuthorRepository {
				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(errors.New("mock error"))
				return authorRepo
			},
			readerRepo: func(ctrl *gomock.Controller) article.ArticleReaderRepository {
				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
				return readerRepo
			},
			wantId:  0,
			wantErr: errors.New("mock error"),
		},
		{
			name: "更新帖子，保存到制作库成功，保存到线上库失败，但重试成功",
			article: domain.Article{
				Id:      1,
				Title:   "测试标题",
				Content: "测试内容",
			},
			authorRepo: func(ctrl *gomock.Controller) article.ArticleAuthorRepository {
				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(nil)
				return authorRepo
			},
			readerRepo: func(ctrl *gomock.Controller) article.ArticleReaderRepository {
				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Store(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(int64(0), errors.New("mock error"))
				readerRepo.EXPECT().Store(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(int64(1), nil)
				return readerRepo
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "更新帖子，保存到制作库成功，保存到线上库失败，重试全部失败",
			article: domain.Article{
				Id:      1,
				Title:   "测试标题",
				Content: "测试内容",
			},
			authorRepo: func(ctrl *gomock.Controller) article.ArticleAuthorRepository {
				authorRepo := artrepomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "测试标题",
					Content: "测试内容",
				}).Return(nil)
				return authorRepo
			},
			readerRepo: func(ctrl *gomock.Controller) article.ArticleReaderRepository {
				readerRepo := artrepomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Store(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "测试标题",
					Content: "测试内容",
				}).Times(3).Return(int64(0), errors.New("mock error"))
				return readerRepo
			},
			wantId:  0,
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authorRepo := tc.authorRepo(ctrl)
			readerRepo := tc.readerRepo(ctrl)
			artSvc := NewArticleServiceV1(nil, authorRepo, readerRepo, &logger.NopLogger{})
			artId, err := artSvc.PublishV1(context.Background(), tc.article)
			assert.Equal(t, tc.wantId, artId)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
