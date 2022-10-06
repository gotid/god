package syncx

import (
	"errors"
	"time"
)

// ErrTimeout 代表借用超时的错误。
var ErrTimeout = errors.New("借用超时")

// TimeoutLimit 用于限时借用。
type TimeoutLimit struct {
	limit Limit
	cond  *Cond
}

// NewTimeoutLimit 返回一个可并发借用n个资源的超时限制。
func NewTimeoutLimit(n int) TimeoutLimit {
	return TimeoutLimit{
		limit: NewLimit(n),
		cond:  NewCond(),
	}
}

// TryBorrow 尝试一次借用。
func (l TimeoutLimit) TryBorrow() bool {
	return l.limit.TryBorrow()
}

// Borrow 在指定的时间内完成借用。
func (l TimeoutLimit) Borrow(timeout time.Duration) error {
	if l.TryBorrow() {
		return nil
	}

	var ok bool
	for {
		timeout, ok = l.cond.WaitWithTimeout(timeout)
		if ok && l.TryBorrow() {
			return nil
		}

		if timeout <= 0 {
			return ErrTimeout
		}
	}
}

// Return 归返一个借用的资源。
func (l TimeoutLimit) Return() error {
	if err := l.limit.Return(); err != nil {
		return err
	}
	l.cond.Signal()
	return nil
}
