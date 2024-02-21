package job

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"time"
	"webookpro/internal/domain"
	"webookpro/internal/service"
	"webookpro/pkg/logger"
)

/*
基于MySQL的分布式任务调度方案
*/

// Executor 执行器
type Executor interface {
	Name() string                                 // 任务执行器的名字
	Exec(ctx context.Context, j domain.Job) error // 真正去执行一个任务
}

// LocalFuncExecutor 我这个执行器就是一个本地方法
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: make(map[string]func(ctx context.Context, j domain.Job) error),
	}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未知任务，你是否注册了？ %s", j.Name)
	}
	return fn(ctx, j)
}

// 调度器
type Scheduler struct {
	execs   map[string]Executor
	svc     service.JobService
	l       logger.Logger
	limiter *semaphore.Weighted // 如果抢占了几十万个任务，会打爆你内存，因此这里加入了一个类似令牌桶的限制
}

func NewScheduler(svc service.JobService, l logger.Logger) *Scheduler {
	return &Scheduler{svc: svc, l: l,
		limiter: semaphore.NewWeighted(200), // 一次最多有个200个任务在运行
		execs:   make(map[string]Executor)}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			// 退出调度循环
			return ctx.Err()
		}
		// 调度前尝试拿一个令牌
		// 如果没拿到会在之类一直阻塞
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		// 拿到了， 直接开始抢占任务
		dbCtx, cancal := context.WithTimeout(ctx, time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancal()
		if err != nil {
			s.l.Error("抢占任务失败", logger.Error(err))
		}
		// 拿到这个任务的执行器
		exec, ok := s.execs[j.Executor]
		if !ok {
			// DEBUG 的时候最好中断
			// 线上就继续
			s.l.Error("未找到对应的执行器",
				logger.String("executor", j.Executor))
			continue
		}
		// 任务有了， 执行器有了，接下来考虑怎么执行这个任务
		// 答案是，你不能阻塞for循环，得开一个go routine异步执行
		go func() {
			defer func() {
				// 在  defer中要释放任务，也要释放令牌
				s.limiter.Release(1)
				err1 := j.CancelFunc()
				if err1 != nil {
					s.l.Error("释放任务失败",
						logger.Error(err1),
						logger.Int64("job_id", j.Id))
				}
			}()
			// 执行任务
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				s.l.Error("任务执行失败", logger.Error(err1))
			}
			// 要考虑下一次调度
			// 即更新 next_time, 这样就能保证我这台节点一直调度这个任务
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("设置下一次执行时间失败", logger.Error(err1))
			}
		}()
	}
}
