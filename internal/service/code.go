package service

import (
	"context"
	"fmt"
	"math/rand"
	"webookpro/internal/repository"
	"webookpro/internal/service/sms"
)

const codeTplId = ""

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type SMSCodeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewSMSCodeService(repo repository.CodeRepository, sms sms.Service) CodeService {
	return &SMSCodeService{
		repo: repo,
		sms:  sms,
	}
}

// Send 发送验证码
func (svc *SMSCodeService) Send(ctx context.Context, biz string, phone string) error {
	code := svc.generateCode(ctx)
	// 1. 将验证码记入缓存
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 2. 成功记入缓存后，发送验证码
	return svc.sms.Send(ctx, codeTplId, []string{code}, phone)
}

// Verify 校验验证码
func (svc *SMSCodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verfiy(ctx, biz, phone, inputCode)
}

// Verify 校验验证码
func (svc *SMSCodeService) generateCode(ctx context.Context) string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
