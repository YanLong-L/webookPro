package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"webookpro/internal/service/sms"
	"webookpro/pkg/ratelimit"
)

type RateLimitSMSService struct {
	sms     sms.Service
	limiter ratelimit.Limiter
}

func NewRateLimitSMSService(sms sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RateLimitSMSService{
		sms:     sms,
		limiter: limiter,
	}

}

func (r *RateLimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// 在这里就可以加上限流功能
	limited, err := r.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候，
		// 可以不限：你的下游很强，业务可用性要求很高，尽量容错策略
		// 包一下这个错误
		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	}
	if limited {
		return errors.New("触发了限流")
	}
	err = r.sms.Send(ctx, tpl, args, numbers...)
	return err
}
