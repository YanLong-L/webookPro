package memory

import (
	"context"
	"fmt"
	"webookpro/internal/service/sms"
)

type Service struct {
}

func NewService() sms.Service {
	return &Service{}
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	fmt.Println(args)
	return nil
}
