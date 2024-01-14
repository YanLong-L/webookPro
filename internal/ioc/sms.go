package ioc

import (
	"webookpro/internal/service/sms"
	"webookpro/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 可以换成别的实现
	return memory.NewService()
}
