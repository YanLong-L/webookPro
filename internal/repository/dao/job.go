package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context, refreshInterval time.Duration) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGORMJobDAO(db *gorm.DB) JobDAO {
	return &GORMJobDAO{
		db: db,
	}
}

func (g *GORMJobDAO) Preempt(ctx context.Context, refreshInterval time.Duration) (Job, error) {
	db := g.db.WithContext(ctx)
	for {
		now := time.Now()
		var j Job
		err := db.WithContext(ctx).
			Where("status = ? AND next_time <= ?", jobStatusWaiting, now).
			Or("utime <= ", time.Now().Add(refreshInterval*3)). // 3个间隔都没续约，我认为你续约失败，是可以抢占的
			First(&j).Error
		if err != nil {
			// 说明没任务执行，直接退出
			return Job{}, err
		}
		// 找到之后，要开始抢占了
		// 抢占： 将该条数据的状态改了
		// 这里强调一下： 为什么需要version, 因为有可能我两个groutine同时走到下面的代码段
		// 第一个goroutine更新成功后，第二个goroutine再去更新的时候
		// id = ? AND version = ? 就找不到了，这样能保证一个任务只会被一个goroutine拿到执行
		res := db.Where("id = ? AND version = ?", j.Id, j.Version).
			Model(&Job{}).Updates(map[string]any{
			"status":  jobStatusRunning,
			"utime":   now,
			"version": j.Version + 1,
		})
		if res.Error != nil {
			return Job{}, err
		}
		if res.RowsAffected == 0 {
			// 抢占失败，被别人抢了，我要继续下一轮
			continue
		}
		// 抢占成功，return job
		return j, nil
	}
}

func (g *GORMJobDAO) Release(ctx context.Context, id int64) error {
	// 这里有一个问题。你要不要检测 status 或者 version?
	// WHERE version = ?
	// 要。
	return g.db.WithContext(ctx).Model(&Job{}).
		Where("id = ? AND status = ?", id, jobStatusRunning).
		Updates(map[string]any{
			"status": jobStatusWaiting,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (g *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).
		Where("id =?", id).Updates(map[string]any{
		"utime": time.Now().UnixMilli(),
	}).Error
}

func (g *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return g.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).Updates(map[string]any{
		"next_time": next.UnixMilli(),
	}).Error
}

func (g *GORMJobDAO) Stop(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).
		Where("id = ?", id).Updates(map[string]any{
		"status": jobStatusPaused,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

type Job struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	Cfg      string
	Executor string
	Name     string `gorm:"unique"`
	Status   int    // 用状态来标记哪些任务可以抢？哪些任务已经被人占着？
	// 另外一个问题，定时任务，我怎么知道，已经到时间了呢？
	NextTime int64  `gorm:"index"` // 下一次被调度的时间
	Cron     string // cron 表达式
	Version  int
	Ctime    int64 // 创建时间，毫秒数
	Utime    int64 // 更新时间，毫秒数
}

const (
	jobStatusWaiting = iota
	// 已经被抢占
	jobStatusRunning
	// 暂停调度
	jobStatusPaused
)
