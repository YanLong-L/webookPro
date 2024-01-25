package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	"webookpro/internal/domain"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Upsert(ctx context.Context, article PublishedArticle) error
	SyncStatus(ctx context.Context, article Article, status domain.ArticleStatus) error
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

// SyncStatus 同步线上库制作库帖子状态
func (d *GORMArticleDAO) SyncStatus(ctx context.Context, article Article, status domain.ArticleStatus) error {
	//TODO implement me
	panic("implement me")
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
	var (
		artId = art.Id
		err   error
	)
	err = d.db.Transaction(func(tx *gorm.DB) error {
		txDAO := NewGORMArticleDAO(tx)
		if art.Id > 0 {
			// 更新制作库
			err = txDAO.UpdateById(ctx, art)
		} else {
			// 新增制作库
			artId, err = txDAO.Insert(ctx, art)
		}
		// upsert 线上库
		return txDAO.Upsert(ctx, PublishedArticle{
			Article: art,
		})
	})
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
			"utime":   article.Utime,
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

type Article struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 长度 1024
	Title   string `gorm:"type=varchar(1024)"`
	Content string `gorm:"type=BLOB"`
	// 如何设计索引
	// 在帖子这里，什么样查询场景？
	// 对于创作者来说，是不是看草稿箱，看到所有自己的文章？
	// SELECT * FROM articles WHERE author_id = 123 ORDER BY `ctime` DESC;
	// 产品经理告诉你，要按照创建时间的倒序排序
	// 单独查询某一篇 SELECT * FROM articles WHERE id = 1
	// 在查询接口，我们深入讨论这个问题
	// - 在 author_id 和 ctime 上创建联合索引
	// - 在 author_id 上创建索引

	// 学学 Explain 命令

	// 在 author_id 上创建索引
	AuthorId int64 `gorm:"index"`
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`
	Status uint8
	Ctime  int64
	Utime  int64
}

type PublishedArticle struct {
	Article
}
