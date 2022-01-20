package fx

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

var (
	// ErrCanceled 是上下文被取消时返回的错误。
	ErrCanceled = context.Canceled

	// ErrTimeout 是上下文截止时间过后返回的错误。
	ErrTimeout = context.DeadlineExceeded
)

// DoOption 是一个自定义 DoWithTimeout 的函数。
type DoOption func() context.Context

// DoWithTimeout 使用超时控制运行 fn。
func DoWithTimeout(fn func() error, timeout time.Duration, opts ...DoOption) error {
	parentCtx := context.Background()
	for _, opt := range opts {
		parentCtx = opt()
	}
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// 创建缓冲区大小为1的通道以避免 goroutine 泄露。
	done := make(chan error, 1)
	panicChan := make(chan interface{}, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				// 附加调用堆栈以避免在不同 goroutine 中丢失
				panicChan <- fmt.Sprintf("%+vn\n%s", p, strings.TrimSpace(string(debug.Stack())))
			}
		}()
		done <- fn()
	}()

	// 处理结果和panic
	select {
	case p := <-panicChan:
		panic(p)
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WithContext 使用指定上下文自定义 DoWithTimeout。
func WithContext(ctx context.Context) DoOption {
	return func() context.Context {
		return ctx
	}
}
