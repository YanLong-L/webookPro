package service

import (
	"context"
	"fmt"
	"math/rand"
	"webookpro/internal/repository"
	"webookpro/internal/service/sms"
)

const codeTplId = ""

type CodeService struct {
	repo *repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo *repository.CodeRepository, sms sms.Service) *CodeService {
	return &CodeService{
		repo: repo,
		sms:  sms,
	}
}

// Send 发送验证码
func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	code := svc.generateCode(ctx)
	// 1. 将验证码记入缓存
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 2. 成功记入缓存后，发送验证码
	return svc.sms.Send(ctx, codeTplId, []string{phone}, code)
}

// Verify 校验验证码
func (svc *CodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verfiy(ctx, biz, phone, inputCode)
}

// Verify 校验验证码
func (svc *CodeService) generateCode(ctx context.Context) string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
