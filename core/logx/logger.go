package logx

import (
	"context"
	"time"
)

type Logger interface {
	// Error 记录一条错误级别的消息。
	Error(...interface{})
	// Errorf 记录一条给定格式的错误级别的消息。
	Errorf(string, ...interface{})
	// Errorv 记录一条 json 编组之后的错误级别的消息。
	Errorv(interface{})
	// Errorw 记录一条给定 LogField 字段的错误级别的消息。
	Errorw(string, ...LogField)

	// Info 记录一条信息级别的消息。
	Info(...interface{})
	// Infof 记录一条给定格式的信息级别的消息。
	Infof(string, ...interface{})
	// Infov 记录一条 json 编组之后的信息级别的消息。
	Infov(interface{})
	// Infow 记录一条给定 LogField 字段的信息级别的消息。
	Infow(string, ...LogField)

	// Slow 记录一条慢执行级别的消息。
	Slow(...interface{})
	// Slowf 记录一条给定格式的慢执行级别的消息。
	Slowf(string, ...interface{})
	// Slowv 记录一条 json 编组之后的慢执行级别的消息。
	Slowv(interface{})
	// Sloww 记录一条给定 LogField 字段的慢执行级别的消息。
	Sloww(string, ...LogField)

	// WithContext 返回具有给定上下文的日志记录器。
	WithContext(ctx context.Context) Logger
	// WithDuration 返回具有给定持续时间的日志记录器。
	WithDuration(duration time.Duration) Logger
}
