package article

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
	"webookpro/internal/domain"
)

type S3ArticleDA0 struct {
	oss *s3.S3
	GORMArticleDAO
	bucket *string
}

func NewS3ArticleDAO(oss *s3.S3, db *gorm.DB) *S3ArticleDA0 {
	return &S3ArticleDA0{
		oss:    oss,
		bucket: ekit.ToPtr[string]("webook-1314583317"),
		GORMArticleDAO: GORMArticleDAO{
			db: db,
		},
	}
}

func (s S3ArticleDA0) Insert(ctx context.Context, article Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s S3ArticleDA0) UpdateById(ctx context.Context, article Article) error {
	//TODO implement me
	panic("implement me")
}

func (s S3ArticleDA0) Sync(ctx context.Context, art Article) (int64, error) {
	// 保存制作库
	// 保存线上库，并且把 content 上传到 OSS
	//
	var (
		id = art.Id
	)
	// 制作库流量不大，并发不高，你就保存到数据库就可以
	// 当然，有钱或者体量大，就还是考虑 OSS
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var err error
		now := time.Now().UnixMilli()
		// 制作库
		txDAO := NewGORMArticleDAO(tx)
		if id == 0 {
			id, err = txDAO.Insert(ctx, art)
		} else {
			err = txDAO.UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		publishArt := PublishedArticle{
			Article{
				Id:       art.Id,
				Title:    art.Title,
				AuthorId: art.AuthorId,
				Status:   art.Status,
				Ctime:    now,
				Utime:    now,
			},
		}
		// 线上库不保存 Content,要准备上传到 OSS 里面
		return tx.Clauses(clause.OnConflict{
			// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  art.Title,
				"utime":  now,
				"status": art.Status,
				// 要参与 SQL 运算的
			}),
		}).Create(&publishArt).Error
	})
	// 说明保存到数据库的时候失败了
	if err != nil {
		return 0, err
	}
	// 接下来就是保存到 OSS 里面
	// 你要有监控，你要有重试，你要有补偿机制
	_, err = s.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      s.bucket,
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return id, err

}

func (s S3ArticleDA0) Upsert(ctx context.Context, article PublishedArticle) error {
	//TODO implement me
	panic("implement me")
}

func (s S3ArticleDA0) SyncStatus(ctx context.Context, article Article, status domain.ArticleStatus) error {
	//TODO implement me
	panic("implement me")
}
