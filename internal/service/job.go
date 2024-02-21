package service

import (
	"context"
	"time"
	"webookpro/internal/domain"
	"webookpro/internal/repository"
	"webookpro/pkg/logger"
)

type JobService interface {
	// Preempt 抢占
	Preempt(ctx context.Context) (domain.Job, error)
	// ResetNextTime
	ResetNextTime(ctx context.Context, j domain.Job) error
}

type CronJobService struct {
	repo            repository.JobRepository
	refreshInterval time.Duration // 刷新utime的时间间隔
	l               logger.Logger
}

func NewCronJobService(repo repository.JobRepository, refreshInterval time.Duration, l logger.Logger) *CronJobService {
	return &CronJobService{repo: repo, refreshInterval: refreshInterval, l: l}
}

func (c *CronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	// 我直接抢占
	job, err := c.repo.Preempt(ctx, c.refreshInterval)
	if err != nil {
		return domain.Job{}, err
	}
	// 任务抢占成功了，如何考虑任务的续约
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			// 不断刷新utime证明自己还活着
			c.Refresh(ctx, job.Id)
		}
	}()
	// 任务抢占了，你要考虑一个释放的问题
	job.CancelFunc = func() error {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return c.repo.Release(ctx, job.Id)
	}
	return job, err
}

// Refresh 刷新job的 utime
func (c *CronJobService) Refresh(ctx context.Context, id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 续约怎么个续法？
	// 更新一下更新时间就可以
	// 比如说我们的续约失败逻辑就是：处于 running 状态，但是更新时间在三分钟以前
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		// 可以考虑立刻重试
		c.l.Error("续约失败",
			logger.Error(err),
			logger.Int64("job_id", id))
	}
}

// ResetNextTime 我们每过一段时间，就刷新这个job的next_time，这样就能确保这个job不会被其他goroutine调度到
func (c *CronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	next := j.NextTime()
	if next.IsZero() {
		// 没有下一次
		return c.repo.Stop(ctx, j.Id)
	}
	return c.repo.UpdateNextTime(ctx, j.Id, next)
}
