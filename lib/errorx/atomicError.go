package errorx

import "sync/atomic"

// AtomicError 定义原子错误。
type AtomicError struct {
	err atomic.Value // error
}

// Set 设置错误。
func (ae *AtomicError) Set(err error) {
	if err != nil {
		ae.err.Store(err)
	}
}

// Load 返回错误。
func (ae *AtomicError) Load() error {
	if err := ae.err.Load(); err != nil {
		return err.(error)
	}

	return nil
}
