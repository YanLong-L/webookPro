package validator

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
	"webookpro/migrator"
	"webookpro/migrator/events"
	"webookpro/pkg/logger"
)

type Validator[T migrator.Entity] struct {
	base      *gorm.DB
	target    *gorm.DB
	l         logger.Logger
	batchSize int
	p         events.Producer
	direction string               // DST or SRC ，判断是用原表去修还是目标表去修
	highLoad  *atomicx.Value[bool] // 是否是高负载
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, l logger.Logger, p events.Producer, direction string) *Validator[T] {
	highLoad := atomicx.NewValueOf[bool](false)
	go func() {
		// 在这里，去查询数据库的状态
		// 你的校验代码不太可能是性能瓶颈，性能瓶颈一般在数据库
		// 你也可以结合本地的 CPU，内存负载来判定
	}()
	return &Validator[T]{base: base, target: target,
		l: l, p: p, direction: direction,
		highLoad: highLoad}
}

// Validate 一次完整的全量校验
func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		v.validateBaseToTarget(ctx)
		return nil
	})
	eg.Go(func() error {
		v.validateTargetToBase(ctx)
		return nil
	})
	return eg.Wait()
}

// validateBaseToTarget  从base到target，进行全量校验
func (v *Validator[T]) validateBaseToTarget(ctx context.Context) {
	offset := -1
	for {
		//
		if v.highLoad.Load() {
			// 可以考虑挂起一段时间
		}
		// 进来就更新 offset，比较好控制
		// 因为后面有很多的 continue 和 return
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		offset++

		// 先从源表中找一条数据 按照id的顺序找
		var src T
		err := v.base.WithContext(dbCtx).Offset(offset).Order("id").First(&src).Error
		cancel()
		switch err {
		case nil:
			// 源表中查到了数据，要去找对应目标表的数据
			var dst T
			// 注意这里如果没有id字段，那就找类似ctime这种排序是和插入顺序一样的字段
			err = v.target.Where("id = ?", src.ID()).First(&dst).Error
			switch err {
			case nil:
				// 目标表数据找到了，开始比较
				// 下面列举几种数据比较的方式
				// 第一种，利用反射去比
				//if reflect.DeepEqual(src, dst) {
				//	// 上报给kafka，数据不一致
				//	v.notify(ctx, src.ID(), events.InconsistentEventTypeNEQ)
				//}
				// 第二种，用自定义的比较逻辑
				if !src.CompareTo(dst) {
					v.notify(ctx, src.ID(), events.InconsistentEventTypeNEQ)
				}
			case gorm.ErrRecordNotFound:
				// 没找到，说明你的目标表中少了数据
				// 没什么说的，上报kafka
				v.notify(ctx, src.ID(), events.InconsistentEventTypeTargetMissing)
			default:
				v.l.Error("查询target数据失败", logger.Error(err))
			}
		case gorm.ErrRecordNotFound:
			// 源表中找不到数据了， 说明都比完了，全量校验结束
			return
		default:
			// 数据库错误
			v.l.Error("校验数据，查询 base 出错",
				logger.Error(err))
			continue
		}
	}
}

// validateTargetToBase 从target到base 进行全量校验
// 因为我们只用源表比对是不够的，因为比对后，base中可能有数据删除，这时候，target中的这部分数据就多了，所以需要再用target去全量校验一遍
func (v *Validator[T]) validateTargetToBase(ctx context.Context) {
	// 先找 target，再找 base，找出 base 中已经被删除的
	// 理论上来说，就是 target 里面一条条找
	// 这这里我们可以采用批量的做法
	offset := -v.batchSize
	for {
		offset = offset + v.batchSize
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		// 取一批目标表中的数据
		var dsts []T
		err := v.target.WithContext(dbCtx).Select("id").
			Offset(offset).Limit(v.batchSize).Order("id").
			Find(&dsts).Error
		cancel()
		if len(dsts) == 0 {
			// 说明目标表数据查完了， 校验结束
			return
		}
		switch err {
		case nil:
			// 目标表查到了一批数据
			// 取到这批数据的id
			ids := slice.Map(dsts, func(idx int, t T) int64 {
				return t.ID()
			})
			var srcs []T
			err = v.base.Where("id IN ?", ids).Find(&srcs).Error
			// 如果在这个地方 原表中没有这部分id，说明目标表这部分id对应的数据就多了，需要删
			if len(srcs) == 0 {
				v.notifyBaseMissing(ctx, ids)
			}
			switch err {
			case nil:
				srcIds := slice.Map(srcs, func(idx int, t T) int64 {
					return t.ID()
				})
				// 计算差集, 也就是，src 里面的没有的数据的id
				diff := slice.DiffSet(ids, srcIds)
				v.notifyBaseMissing(ctx, diff)
			default:
				continue
			}
		default:
			continue
		}
		if len(dsts) < v.batchSize {
			// 没数据了
			return
		}
	}

}

// notify 通知kafka去修数据
func (v *Validator[T]) notify(ctx context.Context, id int64, typ string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	err := v.p.ProduceInconsistentEvent(ctx,
		events.InconsistentEvent{
			ID:        id,
			Direction: v.direction,
			Type:      typ,
		})
	cancel()
	if err != nil {
		// 这又是一个问题
		// 怎么办？
		// 你可以重试，但是重试也会失败，记日志，告警，手动去修
		// 我直接忽略，下一轮修复和校验又会找出来
		v.l.Error("发送数据不一致的消息失败", logger.Error(err))
	}
}

// notifyBaseMissing 通知kafka去修BaseMissing类型的数据
func (v *Validator[T]) notifyBaseMissing(ctx context.Context, ids []int64) {
	for _, id := range ids {
		v.notify(ctx, id, events.InconsistentEventTypeBaseMissing)
	}
}
