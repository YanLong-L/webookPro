package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	"webookpro/internal/domain"
)

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (d *GORMArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var res []PublishedArticle
	err := d.db.WithContext(ctx).
		Where("utime < ?", start.UnixMilli()).
		Order("utime DESC").Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (d *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var art PublishedArticle
	err := d.db.WithContext(ctx).Model(&PublishedArticle{}).Where("id = ?", id).First(&art).Error
	if err != nil {
		return PublishedArticle{}, err
	}
	return art, nil
}

func (d *GORMArticleDAO) GetById(ctx context.Context, artId int64) (Article, error) {
	var art Article
	err := d.db.WithContext(ctx).Model(&Article{}).Where("id = ?", artId).First(&art).Error
	if err != nil {
		return Article{}, err
	}
	return art, err
}

// GetByAuthor 通过authorid 获取 创作者文章列表
func (d *GORMArticleDAO) GetByAuthor(ctx context.Context, authorId int64, offset int, limit int) ([]Article, error) {
	var res []Article
	err := d.db.WithContext(ctx).Model(&Article{}).Where("author_id = ?", authorId).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&res).Error
	return res, err
}

// SyncStatus 同步线上库制作库帖子状态
func (d *GORMArticleDAO) SyncStatus(ctx context.Context, article Article, status domain.ArticleStatus) error {
	now := time.Now().UnixMilli()
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := d.db.Model(&Article{}).Where("id = ? AND author_id = ?", article.Id, article.AuthorId).
			Updates(map[string]any{
				"status": domain.ArticleStatusUnpublished,
				"utime":  now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			// 要么 用户不对 ，要么文章 id不对
			return fmt.Errorf("非法操作 Uid:%s, artId:%s", article.Id, article.AuthorId)
		}
		return d.db.Model(&PublishedArticle{}).Where("id = ? AND author_id = ?", article.Id, article.AuthorId).
			Updates(map[string]any{
				"status": domain.ArticleStatusUnpublished,
				"utime":  now,
			}).Error
	})
	return err
}

// Upsert 线上库upsert
func (d *GORMArticleDAO) Upsert(ctx context.Context, art PublishedArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	// 这个是插入
	// OnConflict 的意思是数据冲突了
	err := d.db.Clauses(clause.OnConflict{
		// SQL 2003 标准
		// INSERT AAAA ON CONFLICT(BBB) DO NOTHING
		// INSERT AAAA ON CONFLICT(BBB) DO UPDATES CCC WHERE DDD
		// 哪些列冲突
		//Columns: []clause.Column{clause.Column{Name: "id"}},
		// 意思是数据冲突，啥也不干
		// DoNothing:
		// 数据冲突了，并且符合 WHERE 条件的就会执行 DO UPDATES
		// Where:

		// MySQL 只需要关心这里
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
			"status":  art.Status,
		}),
	}).Create(&art).Error
	// MySQL 最终的语句 INSERT xxx ON DUPLICATE KEY UPDATE xxx

	// 一条 SQL 语句，都不需要开启事务
	// auto commit: 意思是自动提交

	return err
}

// Sync 帖子发表接口同步线上库和制作库数据
func (d *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	var (
		artId = art.Id
		err   error
	)
	tx := d.db.WithContext(ctx).Begin()
	defer tx.Rollback()
	txDAO := NewGORMArticleDAO(tx)
	if art.Id > 0 {
		// 更新制作库
		err = txDAO.UpdateById(ctx, art)
	} else {
		// 新增制作库
		artId, err = txDAO.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	// upsert 线上库
	publishArt := PublishedArticle{Article: art}
	publishArt.Utime = now
	publishArt.Ctime = now
	err = tx.Clauses(clause.OnConflict{
		// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		}),
	}).Create(&publishArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return artId, err
}

func (d *GORMArticleDAO) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	err := d.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}

func (d *GORMArticleDAO) UpdateById(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	article.Utime = now
	res := d.db.WithContext(ctx).Model(&article).Where("id = ? AND author_id = ?", article.Id, article.AuthorId).
		Updates(map[string]any{
			"title":   article.Title,
			"content": article.Content,
			"utime":   now,
			"status":  article.Status,
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// 此时可能是id 不会或 Authorid 不对
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d",
			article.Id, article.AuthorId)
	}
	return nil
}
