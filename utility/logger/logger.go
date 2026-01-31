package logger

import (
	"context"

	"github.com/LandcLi/LandcLogFace"
)

var globalLogger LandcLogFace.Logger

func InitLogger(config *LandcLogFace.LogConfig) {
	if globalLogger == nil {
		globalLogger = LandcLogFace.GetLoggerWithLogConfig(config)
		globalLogger.Info("使用LogConfig的日志")
	}
}

func InitLoggerWithLogger(logger LandcLogFace.Logger) {
	globalLogger = logger
}

func GetLogger() LandcLogFace.Logger {
	return globalLogger
}
func Debug(ctx context.Context, msg string) {
	globalLogger.WithContext(ctx).Debug(msg)
}
func Info(ctx context.Context, msg string) {
	globalLogger.WithContext(ctx).Info(msg)
}
func Warn(ctx context.Context, msg string) {
	globalLogger.WithContext(ctx).Warn(msg)
}
func Error(ctx context.Context, msg string) {
	globalLogger.WithContext(ctx).Error(msg)
}
func Fatal(ctx context.Context, msg string) {
	globalLogger.WithContext(ctx).Fatal(msg)
}
func Debugf(ctx context.Context, format string, v ...interface{}) {
	globalLogger.WithContext(ctx).Debugf(format, v...)
}
func Infof(ctx context.Context, format string, v ...interface{}) {
	globalLogger.WithContext(ctx).Infof(format, v...)
}
func Warnf(ctx context.Context, format string, v ...interface{}) {
	globalLogger.WithContext(ctx).Warnf(format, v...)
}
func Errorf(ctx context.Context, format string, v ...interface{}) {
	globalLogger.WithContext(ctx).Errorf(format, v...)
}
func Fatalf(ctx context.Context, format string, v ...interface{}) {
	globalLogger.WithContext(ctx).Fatalf(format, v...)
}
