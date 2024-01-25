package article

import (
	"context"
	"gorm.io/gorm"
	"webookpro/internal/domain"
	"webookpro/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncV1(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx context.Context, art domain.Article, status domain.ArticleStatus) error
}

type CachedArticleRepository struct {
	dao article.ArticleDAO

	// SyncV1  操作两个 DAO
	authorDAO ArticleAuthorRepository
	readerDAO ArticleReaderRepository
	db        *gorm.DB
}

func NewCachedArticleRepository(dao article.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

// SyncStatus 同步帖子状态
func (r *CachedArticleRepository) SyncStatus(ctx context.Context, art domain.Article, status domain.ArticleStatus) error {
	return r.dao.SyncStatus(ctx, r.domainToEntity(art), status)
}

// Sync 帖子发表接口同步数据
func (r *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	return r.dao.Sync(ctx, r.domainToEntity(article))
}

// SyncV1 尝试在repo上控制事务，要引入两个dao
func (r *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	tx := r.db.WithContext(ctx).Begin()
	defer tx.Rollback()
	if tx.Error != nil {
		return 0, tx.Error
	}
	authorDAO := article.NewArticleAuthorDAO(tx)
	readerDAO := article.NewArticleReaderDAO(tx)
	var (
		artId = art.Id
		err   error
	)
	if art.Id > 0 {
		// 更新制作库
		err = authorDAO.UpdateById(ctx, r.domainToEntity(art))
	} else {
		// 新增制作库
		artId, err = authorDAO.Insert(ctx, r.domainToEntity(art))
	}
	if err != nil {
		// 执行有问题，要回滚
		//tx.Rollback()
		return artId, err
	}
	// upsert 线上库
	err = readerDAO.Upsert(ctx, article.PublishedArticle{
		Article: r.domainToEntity(art),
	})
	if err != nil {
		return artId, err
	}
	tx.Commit()
	return artId, err
}

func (r *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return r.dao.Insert(ctx, r.domainToEntity(article))
}

func (r *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	return r.dao.UpdateById(ctx, r.domainToEntity(article))
}

func (r *CachedArticleRepository) domainToEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}
