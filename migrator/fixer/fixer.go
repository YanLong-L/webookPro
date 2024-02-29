package fixer

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"webookpro/migrator"
	"webookpro/migrator/events"
)

type Fixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

// Fix 最一了百了的写法 直接用base覆盖target
func (f *Fixer[T]) Fix(ctx context.Context, evt events.InconsistentEvent) error {
	var t T
	err := f.base.WithContext(ctx).
		Where("id =?", evt.ID).First(&t).Error
	switch err {
	case nil:
		// base 有数据
		// 修复数据的时候，可以考虑增加 WHERE base.utime >= target.utime
		// utime 用不了，就看有没有version 之类的，或者能够判定数据新老的
		return f.target.WithContext(ctx).
			Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
	case gorm.ErrRecordNotFound:
		// base 没了
		return f.target.WithContext(ctx).
			Where("id=?", evt.ID).Delete(&t).Error
	default:
		return err
	}
}

// FixV1 最初版本的Fix
// 这里要抓住一个核心，就base在校验时候的数据，到你修复的时候可能就变了
// 因此在修复之前
func (f *Fixer[T]) FixV1(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	// 目标表缺失数据
	case events.InconsistentEventTypeTargetMissing:
		// target 要插入
		// 插入之前，我先查一下我base的数据，万一已经被删了呢
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// base中这条也没有了，啥都不用做
			return nil
		case nil:
			// base中这条依然在，那target中就得插入这条数据才能完成修复
			return f.target.Create(&t).Error
		default:
			return err
		}
	// 目标表与源表数据不一致
	case events.InconsistentEventTypeNEQ:
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// 我想用base修，结果你源表数据已经删了，那我目标表也得删
			return f.target.Where("id = ?", evt.ID).Delete(&t).Error
		case nil:
			return f.target.Updates(&t).Error
		default:
			return err
		}
	// 目标表多数据
	case events.InconsistentEventTypeBaseMissing:
		return f.target.Where("id = ?", evt.ID).Delete(new(T)).Error
	default:
		return errors.New("未知的不一致类型")
	}
}

// FixV2  InconsistentEventTypeTargetMissing 和  InconsistentEventTypeNEQ 可以合并
func (f *Fixer[T]) FixV2(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeTargetMissing,
		events.InconsistentEventTypeNEQ:
		// 这边要插入
		var t T
		err := f.base.WithContext(ctx).
			Where("id =?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// base 也删除了这条数据
			return f.target.WithContext(ctx).
				Where("id=?", evt.ID).Delete(new(T)).Error
		case nil:
			return f.target.Clauses(clause.OnConflict{
				// 这边要更新全部列
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
		default:
			return err
		}
		// 这边要更新
	case events.InconsistentEventTypeBaseMissing:
		return f.target.WithContext(ctx).
			Where("id=?", evt.ID).Delete(new(T)).Error
	default:
		return errors.New("未知的不一致类型")
	}
}
