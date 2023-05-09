package logger

import "github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/log"

var (
	defaultLogger log.ILogger
)

// SetLogger set global internal logger
func SetLogger(logger log.ILogger) {
	defaultLogger = logger
}

func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Fatal(format, args...)
}
