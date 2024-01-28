package article

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
	"webookpro/internal/domain"
	"webookpro/internal/repository/cache"
	"webookpro/internal/repository/dao/article"
	"webookpro/pkg/logger"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncV1(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx context.Context, art domain.Article, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao   article.ArticleDAO
	cache cache.ArticleCache
	l     logger.Logger

	// SyncV1  操作两个 DAO
	authorDAO ArticleAuthorRepository
	readerDAO ArticleReaderRepository
	db        *gorm.DB
}

func NewCachedArticleRepository(dao article.ArticleDAO, cache cache.ArticleCache, l logger.Logger) ArticleRepository {
	return &CachedArticleRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

// List 创作者文章列表
func (r *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 在列表查询页，集成一些查询方案
	// 先查缓存
	// 如果是访问第一页，并且limit <= 100的时候 ，我直接缓存
	if offset == 0 && limit <= 100 {
		// 直接走缓存
		res, err := r.cache.GetFirstPage(ctx, uid)
		if err == nil {
			// 业务预加载，我预测用户加载完第一页数据后，极有可能会访问第一条，所以我直接预加载
			go func() {
				r.PreCache(ctx, res[0])
			}()
			return res[:limit], err
		}
	}
	var res []domain.Article
	artList, err := r.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return res, err
	}
	res = slice.Map[article.Article, domain.Article](artList, func(idx int, src article.Article) domain.Article {
		return r.entitytoDomain(artList[idx])
	})
	// 查完数据库，同步到缓存
	go func() {
		err := r.cache.SetFirstPage(ctx, uid, res)
		r.l.Error("回写缓存失败", logger.Error(err))
		r.PreCache(ctx, res[0])
	}()
	return res, err
}

// SyncStatus 同步帖子状态
func (r *CachedArticleRepository) SyncStatus(ctx context.Context, art domain.Article, status domain.ArticleStatus) error {
	return r.dao.SyncStatus(ctx, r.domainToEntity(art), status)
}

// Sync 帖子发表接口同步数据
func (r *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	id, err := r.dao.Sync(ctx, r.domainToEntity(article))
	if err == nil {
		err = r.cache.DelFirstPage(ctx, article.Author.Id)
		if err != nil {
			r.l.Error("删除第一页帖子缓存失败", logger.Error(err))
		}
		err = r.cache.SetPub(ctx, article)
		if err != nil {
			r.l.Error("设置线上库帖子缓存失败", logger.Error(err))
		}
		return id, err
	}
	return id, err
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

func (r *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		err := r.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			r.l.Error("清空第一页帖子缓存失败", logger.Error(err))
		}
	}()
	return r.dao.Insert(ctx, r.domainToEntity(art))
}

func (r *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	defer func() {
		// 清空缓存
		err := r.cache.DelFirstPage(ctx, article.Author.Id)
		if err != nil {
			r.l.Error("清空第一页帖子缓存失败", logger.Error(err))
		}
	}()
	return r.dao.UpdateById(ctx, r.domainToEntity(article))
}

// PreCache 业务预加载，提前加载好第一条数据
func (r *CachedArticleRepository) PreCache(ctx context.Context, art domain.Article) {
	err := r.cache.Set(ctx, art)
	if err != nil {
		r.l.Error("业务预加载缓存失败", logger.Error(err))
	}
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

func (repo *CachedArticleRepository) entitytoDomain(art article.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
}
