package service

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
	"webookpro/internal/domain"
	events "webookpro/internal/events/article"
	"webookpro/internal/repository/article"
	"webookpro/pkg/logger"
)

type ArticleServcie interface {
	Store(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	PublishV1(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx *gin.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo article.ArticleRepository

	// 引入两个repo
	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository

	// 引入logger
	l        logger.Logger
	producer events.Producer
}

func NewArticleService(repo article.ArticleRepository,
	producer events.Producer,
	l logger.Logger) ArticleServcie {
	return &articleService{
		repo:     repo,
		l:        l,
		producer: producer,
	}
}

// GetPublishedById 获取线上库帖子详情
func (s *articleService) GetPublishedById(ctx *gin.Context, artId int64, uid int64) (domain.Article, error) {
	art, err := s.repo.GetPublishedById(ctx, artId)
	if err == nil {
		go func() {
			er := s.producer.ProduceReadEvent(
				ctx,
				events.ReadEvent{
					// 即便你的消费者要用 art 的里面的数据，
					// 让它去查询，你不要在 event 里面带
					Uid: uid,
					Aid: artId,
				})
			if er == nil {
				s.l.Error("发送读者阅读事件失败")
			}
		}()
	}
	return art, err
}

// GetById 获取创作者文章详情
func (s *articleService) GetById(ctx context.Context, artId int64) (domain.Article, error) {
	return s.repo.GetById(ctx, artId)
}

// List 创作者文章列表
func (s *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return s.repo.List(ctx, uid, offset, limit)
}

// Withdraw 撤回帖子发表
func (s *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	return s.repo.SyncStatus(ctx, art, domain.ArticleStatusUnpublished)
}

// v1: NewArticleServiceV1 在serice层操作 author和 reader 两个repo
func NewArticleServiceV1(repo article.ArticleRepository, author article.ArticleAuthorRepository,
	reader article.ArticleReaderRepository, l logger.Logger) ArticleServcie {
	return &articleService{
		repo:   repo,
		reader: reader,
		author: author,
		l:      l,
	}
}

// Store 制作库编辑一篇文章并保存
func (s *articleService) Store(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusUnpublished
	if article.Id > 0 {
		err := s.repo.Update(ctx, article)
		return article.Id, err
	}
	id, err := s.repo.Create(ctx, article)
	return id, err
}

// Publish 发表帖子
func (s *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	artId, err := s.repo.Sync(ctx, article)
	return artId, err
}

// PublishV1 V1: publish层操作两个repo
func (s *articleService) PublishV1(ctx context.Context, article domain.Article) (int64, error) {
	var (
		artId = article.Id
		err   error
	)
	if artId > 0 {
		// 制作库更新
		err = s.author.Update(ctx, article)
	} else {
		// 制作库新建
		artId, err = s.author.Create(ctx, article)
	}
	if err != nil { // 制作库都保存失败了，直接返回
		return 0, err
	}
	// 保证线上库和制作库的id是一样的
	article.Id = artId
	// 线上库 upsert  有就更新，没有就新建
	for i := 0; i < 3; i++ {
		artId, err = s.reader.Store(ctx, article)
		if err == nil {
			break
		}
		s.l.Error(fmt.Sprintf("保存到线上库失败，重试第%d次", i),
			logger.Int64("art_id", article.Id),
			logger.Error(err))
		time.Sleep(time.Second * time.Duration(i))
	}
	s.l.Error("保存到线上库失败，重试全部失败",
		logger.Int64("art_id", article.Id),
		logger.Error(err))

	return artId, err
}
