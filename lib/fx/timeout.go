package fx

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

var (
	// ErrCanceled 代表上下文取消的错误。
	ErrCanceled = context.Canceled
	// ErrTimeout 代表上下文过期的错误。
	ErrTimeout = context.DeadlineExceeded
)

// DoOption 自定义 DoWithTimeout。
type DoOption func() context.Context

// DoWithTimeout 带超时控制的函数执行方法。
func DoWithTimeout(fn func() error, timeout time.Duration, opts ...DoOption) error {
	parentCtx := context.Background()
	for _, opt := range opts {
		parentCtx = opt()
	}
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// 创建缓冲为1的通道，以防协程泄露
	done := make(chan error, 1)
	panicChan := make(chan interface{}, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				// 挂接调用堆栈，以免在不同的协程中丢失
				panicChan <- fmt.Sprintf("%+v\n\n%s", p, strings.TrimSpace(string(debug.Stack())))
			}
		}()
		done <- fn()
	}()

	select {
	case p := <-panicChan:
		panic(p)
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WithContext 自定义 DoWithTimeout 调用的上下文。
func WithContext(ctx context.Context) DoOption {
	return func() context.Context {
		return ctx
	}
}
