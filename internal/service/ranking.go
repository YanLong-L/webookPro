package service

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
	"webookpro/internal/domain"
	"webookpro/internal/repository"
)

type RankingService interface {
	TopN(ctx context.Context) error
	//TopN(ctx context.Context, n int64) error
	//TopN(ctx context.Context, n int64) ([]domain.Article, error)
}

type BatchRankingService struct {
	artSvc    ArticleService
	intrSvc   InteractiveService
	repo      repository.RankingRepository
	batchSize int
	n         int
	// scoreFunc 计算文章的热点分数，切不能返回负数
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc InteractiveService) *BatchRankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(float64(sec+2), 1.5)
		},
	}
}

func (svc *BatchRankingService) TopN(ctx context.Context) error {
	// 计算topN数据
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	// 将topN数据缓存
	return svc.repo.ReplaceTopN(ctx, arts)
}

func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	type Score struct {
		art   domain.Article
		score float64
	}
	// 这里可以用非并发安全
	// 初始化一个优先级队列
	topN := queue.NewConcurrentPriorityQueue[Score](svc.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})
	// 在这里开始分批往这个队列中放数据
	// 我只取7天内的数据
	now := time.Now()
	offset := 0
	for {
		// 先拿一批文章
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})
		// 获取这批文章的点赞数据
		intrs, err := svc.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		// 将这批文章加入进优先级队列
		for _, art := range arts {
			// 拿到当前文章的互动信息（点赞，收藏。。）
			intr := intrs[art.Id]
			// 计算当前文章的热度得分
			score := svc.scoreFunc(art.Utime, intr.LikeCnt)
			// 先尝试把这篇文章加入队列，如果队列已经满了会有err
			err := topN.Enqueue(Score{
				art:   art,
				score: score,
			})
			// 这种写法要求 topN 已经满了
			if err == queue.ErrOutOfCapacity {
				val, _ := topN.Dequeue()
				if val.score < score {
					err = topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				} else {
					_ = topN.Enqueue(val)
				}
			}

		}
		// 此时，当前批已经处理完了
		// 我怎么知道要不要进入下一批，我怎么知道还没有没有
		if len(arts) < svc.batchSize ||
			now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			// 如果我这一批都没取够，我肯定没有下一批了
			// 或者我都取到7天之前的数据了，肯定也不用再取了
			break
		}
		// 更新offset
		offset = offset + svc.batchSize
	}
	// 得出结果
	res := make([]domain.Article, svc.n)
	for i := svc.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			break
		}
		res[i] = val.art
	}
	return res, nil
}
