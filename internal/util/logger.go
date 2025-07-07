package util

import (
	"context"
	"log"
	"os"
	"time"
)

// Logger 结构化日志接口
type Logger interface {
	Info(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, err error, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Debug(ctx context.Context, msg string, fields ...Field)
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// logger 日志实现
type logger struct {
	*log.Logger
}

// NewLogger 创建日志实例
func NewLogger() Logger {
	return &logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info 信息日志
func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log("INFO", msg, nil, fields...)
}

// Error 错误日志
func (l *logger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	l.log("ERROR", msg, err, fields...)
}

// Warn 警告日志
func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log("WARN", msg, nil, fields...)
}

// Debug 调试日志
func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log("DEBUG", msg, nil, fields...)
}

// log 内部日志方法
func (l *logger) log(level, msg string, err error, fields ...Field) {
	// 构建结构化日志
	logData := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level,
		"message":   msg,
	}

	// 添加错误信息
	if err != nil {
		logData["error"] = err.Error()
	}

	// 添加自定义字段
	for _, field := range fields {
		logData[field.Key] = field.Value
	}

	// 输出日志
	l.Printf("[%s] %s", level, msg)
	if len(fields) > 0 || err != nil {
		l.Printf("  Fields: %+v", logData)
	}
}

// 全局日志实例
var GlobalLogger Logger

func init() {
	GlobalLogger = NewLogger()
}
