package logx

import (
	"context"
	"time"
)

type Logger interface {
	// Debug 记录一条调试级别的消息。
	Debug(...any)
	// Debugf 记录一条调试级别的消息。
	Debugf(string, ...any)
	// Debugv 记录一条调试级别的消息。
	Debugv(any)
	// Debugw 记录一条调试级别的消息。
	Debugw(string, ...LogField)

	// Error 记录一条错误级别的消息。
	Error(...any)
	// Errorf 记录一条错误级别的消息。
	Errorf(string, ...any)
	// Errorv 记录一条错误级别的消息
	Errorv(any)
	// Errorw 记录一条错误级别的消息
	Errorw(string, ...LogField)

	// Info 记录一条信息级别的消息。
	Info(...any)
	// Infof 记录一条信息级别的消息。
	Infof(string, ...any)
	// Infov 记录一条信息级别的消息。
	Infov(any)
	// Infow 记录一条信息级别的消息。
	Infow(string, ...LogField)

	// Slow 记录一条慢执行级别的消息。
	Slow(...any)
	// Slowf 记录一条慢执行级别的消息。
	Slowf(string, ...any)
	// Slowv 记录一条慢执行级别的消息。
	Slowv(any)
	// Sloww 记录一条慢执行级别的消息。
	Sloww(string, ...LogField)

	// WithContext 返回具有给定上下文的日志记录器。
	WithContext(ctx context.Context) Logger
	// WithDuration 返回具有给定持续时间的日志记录器。
	WithDuration(duration time.Duration) Logger
	// WithCallerSkip 返回具有给定调用者跳跃级别的日志记录器。
	WithCallerSkip(skip int) Logger
	// WithFields 返回具有给定字段的日志记录器。
	WithFields(fields ...LogField) Logger
}
