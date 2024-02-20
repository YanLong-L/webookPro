package job

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
	"webookpro/internal/service"
	"webookpro/pkg/logger"
)

type RankingJob struct {
	svc       service.RankingService // 定时运行这个服务
	timeout   time.Duration          // 分布式锁的过期时间
	client    *rlock.Client
	key       string
	l         logger.Logger
	lock      *rlock.Lock
	localLock *sync.Mutex
}

func NewRankingJob(svc service.RankingService,
	timeout time.Duration,
	client *rlock.Client,
	l logger.Logger) *RankingJob {
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		client:    client,
		key:       "rlock:cron_job:ranking",
		l:         l,
		localLock: &sync.Mutex{},
	}
}

func (r RankingJob) Name() string {
	return "ranking"
}

// Run 每三分钟调度一次
func (r RankingJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	if r.lock == nil {
		// 说明你没拿到锁，你得尝试拿锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 试着拿锁
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			// 说明没拿到锁，极大概率是别人持有了锁
			return nil
		}
		r.lock = lock
		// 如果我拿到了这个锁，我怎么保证我一直拿着这个锁？
		// 答案是，自动续约，当我锁还没过期的时候，我就再续约一次
		go func() {
			r.localLock.Lock()
			defer r.localLock.Unlock()
			// 自动续约
			err1 := lock.AutoRefresh(r.timeout/2, time.Second)
			if err1 != nil {
				// 在这里，如果续约失败了怎么办？
				// 打个日志，争取下一次继续抢锁
				r.l.Error("续约失败", logger.Error(err1))
			}
			r.lock = nil
		}()
	}
	// 锁拿完了，调用 ranking服务记录topn数据
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
