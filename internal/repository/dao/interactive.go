package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

//go:generate mockgen -source=./interactive.go -package=daomocks -destination=mocks/interactive.mock.go InteractiveDAO
type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error)
	DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}
}

func (dao GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDAO := NewGORMInteractiveDAO(tx)
		for i := range bizs {
			err := txDAO.IncrReadCnt(ctx, bizs[i], ids[i])
			if err != nil {
				// 记个日志就拉到
				// 也可以 return err
				return err
			}
		}
		return nil
	})
}

// IncrReadCnt 增加阅读计数
func (dao GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("read_cnt + 1"),
			"utime":    time.Now().UnixMilli(),
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

func (dao GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	// 要更新对于该资源的点赞数
	// 要记录用户对该资源点过赞
	// 以上操作要在一个事务内完成
	// 两个操作都是类似upsert的语义
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// upsert 资源交互表中的点赞数
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("like_cnt + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   bizId,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
		if err != nil {
			return err
		}
		// upsert 用户给该资源点过赞
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"status": 1,
				"utime":  now,
			}),
		}).Create(&UserLikeBiz{
			Biz:    biz,
			BizId:  bizId,
			Uid:    uid,
			Ctime:  now,
			Utime:  now,
			Status: 1,
		}).Error

	})
}

func (dao GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).
		Where("biz=? AND biz_id = ? AND uid = ? AND status = ?",
			biz, bizId, uid, 1).First(&res).Error
	return res, err
}

func (dao GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	now := time.Now().UnixMilli()
	// 减少资源交互表中对应数据的点赞数
	// 软删除用户点赞表中的一条数据
	// 以上从操作在一个事务内完成
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 软删除用户点赞表中的一条数据
		err := tx.Model(&UserLikeBiz{}).
			Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).
			Updates(map[string]any{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).Where("biz = ? AND biz_id = ?", biz, bizId).
			Updates(map[string]any{
				"utime":    now,
				"like_cnt": gorm.Expr("like_cnt - 1"),
			}).Error
	})
}

// Get 获取点赞，阅读，收藏数
func (dao GORMInteractiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ?", biz, bizId).
		First(&res).Error
	return res, err
}

func (dao GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Utime = now
	cb.Ctime = now
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := dao.db.WithContext(ctx).Create(&cb).Error
		if err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"collect_cnt": gorm.Expr("`collect_cnt`+1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
			Biz:        cb.Biz,
			BizId:      cb.BizId,
		}).Error
	})
}

func (dao GORMInteractiveDAO) GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).
		Where("biz=? AND biz_id = ? AND uid = ?", biz, bizId, uid).First(&res).Error
	return res, err
}

// Interactive 对于一个biz的交互信息表
type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_id_type"`
	Biz        string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

// UserLikeBiz 用户点赞的某个东西
// 来判定你有没有点赞 WHERE uid = ? AND biz_id = ? AND biz = ?
type UserLikeBiz struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Biz    string `gorm:"uniqueIndex:uid_biz_id_type;type:varchar(128)"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Ctime  int64
	Utime  int64
	Status uint8 // 0-代表删除，1 代表有效
}

// Collection 收藏夹
type Collection struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Name  string `gorm:"type=varchar(1024)"`
	Uid   int64  `gorm:""`
	Ctime int64
	Utime int64
}

// UserCollectionBiz 用户收藏的东西
type UserCollectionBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Cid   int64  `gorm:"index"` //收藏夹 ID 作为关联关系中的外键，我们这里需要索引
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	Uid   int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Ctime int64
	Utime int64
}
