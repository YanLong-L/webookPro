package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webookpro/internal/service/sms"
)

type PrometheusSMSService struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewPrometheusSMSService(svc sms.Service) *PrometheusSMSService {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "sms_resp_time",
		Help:      "统计 SMS 服务的性能数据",
	}, []string{"biz"})
	prometheus.MustRegister(vector)
	return &PrometheusSMSService{
		svc:    svc,
		vector: vector,
	}
}

func (p *PrometheusSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.vector.WithLabelValues(tpl).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, tpl, args, numbers...)
}
