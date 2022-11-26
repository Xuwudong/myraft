package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

const (
	TraceIDKey = "trace_id"
)

// Logger Logrus
type Logger = logrus.Logger

// Entry Logrus entry
type Entry = logrus.Entry

// Hook 定义日志钩子别名
type Hook = logrus.Hook

// StandardLogger 获取标准日志
func StandardLogger() *Logger {
	return logrus.StandardLogger()
}

// SetLevel 设定日志级别
func SetLevel(level int) {
	logrus.SetLevel(logrus.Level(level))
}

// SetFormatter 设定日志输出格式
func SetFormatter(format string) {
	switch format {
	case "json":
		logrus.SetFormatter(new(logrus.JSONFormatter))
	default:
		logrus.SetFormatter(new(logrus.TextFormatter))
	}
}

// AddHook 增加日志钩子
func AddHook(hook Hook) {
	logrus.AddHook(hook)
}

// FromTraceIDContext 从上下文中获取跟踪ID
func FromTraceIDContext(ctx context.Context) string {
	v := ctx.Value(TraceIDKey)
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// WithContext Use context create entry
func WithContext(ctx context.Context) *Entry {
	if ctx == nil {
		ctx = context.Background()
	}

	fields := map[string]interface{}{}

	if v := FromTraceIDContext(ctx); v != "" {
		fields[TraceIDKey] = v
	}

	return logrus.WithContext(ctx).WithFields(fields)
}

var (
	Tracef = logrus.Tracef
	Debugf = logrus.Debugf
	Infof  = logrus.Infof
	Warnf  = logrus.Warnf
	Errorf = logrus.Errorf
	Fatalf = logrus.Fatalf
	Fatal  = logrus.Fatal
	Panicf = logrus.Panicf
	Printf = logrus.Printf
	Info   = logrus.Info
	Debug  = logrus.Debug
	Error  = logrus.Error
)
