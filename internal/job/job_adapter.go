package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webookpro/pkg/logger"
)

type JobAdapter struct {
	j Job
	l logger.Logger
	p prometheus.Summary
}

func NewJobAdapter(j Job, l logger.Logger) *JobAdapter {
	p := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "cron_job",
		ConstLabels: map[string]string{
			"name": j.Name(),
		},
	})
	prometheus.MustRegister(p)
	return &JobAdapter{
		j: j,
		l: l,
		p: p,
	}
}

func (r *JobAdapter) Run() {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		r.p.Observe(float64(duration))
	}()
	err := r.j.Run()
	if err != nil {
		r.l.Error("运行任务失败", logger.Error(err),
			logger.String("job", r.j.Name()))
	}
}
