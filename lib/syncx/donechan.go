package syncx

import (
	"github.com/gotid/god/lib/lang"
	"sync"
)

// DoneChan 用作可以多次关闭并等待完成的通道。
type DoneChan struct {
	done chan lang.PlaceholderType
	once sync.Once
}

// NewDoneChan 返回一个 DoneChan。
func NewDoneChan() *DoneChan {
	return &DoneChan{
		done: make(chan lang.PlaceholderType),
	}
}

// Close 关闭通道。多次关闭是安全的。
func (dc *DoneChan) Close() {
	dc.once.Do(func() {
		close(dc.done)
	})
}

// Done 返回可以在dc关闭时被通知的通道。
func (dc *DoneChan) Done() chan lang.PlaceholderType {
	return dc.done
}
