package logger

import (
	"context"

	"github.com/LandcLi/landc-go/log/facade"
)

var globalLogger facade.Logger

func InitLogger(config *facade.LogConfig) {
	if globalLogger == nil {
		globalLogger = facade.GetLoggerWithLogConfig(config)
	}
}

func InitLoggerWithLogger(logger facade.Logger) {
	globalLogger = logger
}

func GetLogger() facade.Logger {
	return globalLogger
}

func Debug(ctx context.Context, msg string) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Debug(msg)
	}
}
func Info(ctx context.Context, msg string) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Info(msg)
	}
}
func Warn(ctx context.Context, msg string) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Warn(msg)
	}
}
func Error(ctx context.Context, msg string) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Error(msg)
	}
}
func Fatal(ctx context.Context, msg string) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Fatal(msg)
	}
}
func Debugf(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Debugf(format, v...)
	}
}
func Infof(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Infof(format, v...)
	}
}
func Warnf(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Warnf(format, v...)
	}
}
func Errorf(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Errorf(format, v...)
	}
}
func Fatalf(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.WithContext(ctx).Fatalf(format, v...)
	}
}
