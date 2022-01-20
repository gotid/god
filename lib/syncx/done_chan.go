package syncx

import (
	"sync"

	"github.com/gotid/god/lib/lang"
)

// DoneChan 可多次关闭和等待完成的通道（优雅通知关闭）
type DoneChan struct {
	done chan lang.PlaceholderType
	once sync.Once
}

// NewDoneChan 返回一个 DoneChan
func NewDoneChan() *DoneChan {
	return &DoneChan{
		done: make(chan lang.PlaceholderType),
	}
}

// Close 关闭dc通道，可安全地多次调用。
func (dc *DoneChan) Close() {
	dc.once.Do(func() {
		close(dc.done)
	})
}

// Done 在关闭dc通道时返回一个可被通知的通道
func (dc *DoneChan) Done() chan lang.PlaceholderType {
	return dc.done
}
