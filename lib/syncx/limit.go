package syncx

import (
	"errors"
	"github.com/gotid/god/lib/lang"
)

// ErrLimitReturn 代表归还的资源多于借用的资源。
var ErrLimitReturn = errors.New("归还的资源多于借用资源了")

// Limit 用于限制并发请求数。
type Limit struct {
	pool chan lang.PlaceholderType
}

// NewLimit 返回一个可并发借用n个资源的限制。
func NewLimit(n int) Limit {
	return Limit{
		pool: make(chan lang.PlaceholderType, n),
	}
}

// Borrow 在阻塞模式下，从 Limit 借用一个资源。
// 借用1次，放1个占位符。
func (l Limit) Borrow() {
	l.pool <- lang.Placeholder
}

// TryBorrow 在阻塞模式下，尝试从 Limit 借用一个资源。
// 成功返回 true，反之返回 false。
func (l Limit) TryBorrow() bool {
	select {
	case l.pool <- lang.Placeholder:
		return true
	default:
		return false
	}
}

// Return 归还借用的资源。当多次归还时返回错误。
// 归还1个，则从池中释放1个。
func (l Limit) Return() error {
	select {
	case <-l.pool:
		return nil
	default:
		return ErrLimitReturn
	}
}
