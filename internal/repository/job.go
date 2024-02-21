package repository

import (
	"context"
	"time"
	"webookpro/internal/domain"
	"webookpro/internal/repository/dao"
)

type JobRepository interface {
	// 抢占
	Preempt(ctx context.Context, refreshInterval time.Duration) (domain.Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type CronJobRepository struct {
	dao dao.JobDAO
}

func NewCronJobRepository(dao dao.JobDAO) JobRepository {
	return &CronJobRepository{
		dao: dao,
	}
}

func (c *CronJobRepository) Release(ctx context.Context, id int64) error {
	return c.dao.Release(ctx, id)
}

func (c *CronJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return c.dao.UpdateUtime(ctx, id)
}

func (c *CronJobRepository) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return c.dao.UpdateNextTime(ctx, id, next)
}

func (c *CronJobRepository) Stop(ctx context.Context, id int64) error {
	return c.dao.Stop(ctx, id)
}

func (c *CronJobRepository) Preempt(ctx context.Context, refreshInterval time.Duration) (domain.Job, error) {
	job, err := c.dao.Preempt(ctx, refreshInterval)
	if err != nil {
		return domain.Job{}, err
	}
	return domain.Job{
		Cfg:      job.Cfg,
		Id:       job.Id,
		Name:     job.Name,
		Executor: job.Executor,
	}, nil
}
