package logger

//type Logger interface {
//	Debug(msg string, args ...any)
//	Info(msg string, args ...any)
//	Warn(msg string, args ...any)
//	Error(msg string, args ...any)
//}
//
//func LoggerExample() {
//	var l Logger
//	phone := "152xxxx1234"
//	l.Info("用户未注册，手机号码是 %s", phone)
//}

type Logger interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}

type Field struct {
	Key   string
	Value any
}
