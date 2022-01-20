package syncx

import (
	"errors"

	"github.com/gotid/god/lib/lang"
)

var ErrLimitReturn = errors.New("请求限制")

// Limit 控制并发请求。
type Limit struct {
	pool chan lang.PlaceholderType
}

// NewLimit 新建指定并发数的限制资源池。
func NewLimit(n int) Limit {
	return Limit{
		pool: make(chan lang.PlaceholderType, n),
	}
}

// Borrow 在阻塞模式下借用一个资源。
func (l Limit) Borrow() {
	l.pool <- lang.Placeholder
}

// Return 归还一个可借资源；否则返回错误。
func (l Limit) Return() error {
	select {
	case <-l.pool:
		return nil
	default:
		return ErrLimitReturn
	}
}

// TryBorrow 在非阻塞模式下借用一个资源并返回真，反之返回假。
func (l Limit) TryBorrow() bool {
	select {
	case l.pool <- lang.Placeholder:
		return true
	default:
		return false
	}
}
