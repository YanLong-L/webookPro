package startup

import "webookpro/pkg/logger"

func InitLog() logger.Logger {
	return logger.NewNoOpLogger()
}
