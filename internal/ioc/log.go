package ioc

import (
	"webookpro/pkg/logger"
)

func InitLogger() logger.Logger {
	//l, err := zap.NewDevelopment()
	//if err != nil {
	//	panic(err)
	//}
	return logger.NewNoOpLogger()
}
