package syncx

import (
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/timex"
	"time"
)

// Cond 用于等待一些条件。
type Cond struct {
	signal chan lang.PlaceholderType
}

// Wait 等待一些信号。
func (c *Cond) Wait() {
	<-c.signal
}

// WaitWithTimeout 带有超时时间的等待。返回剩余等待时长及是否还有时间。
func (c *Cond) WaitWithTimeout(timeout time.Duration) (time.Duration, bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	begin := timex.Now()
	select {
	case <-c.signal:
		elapsed := timex.Since(begin)
		remainTimeout := timeout - elapsed
		return remainTimeout, true
	case <-timer.C:
		return 0, false
	}
}

// Signal 唤醒一个等待该条件的goroutine。
func (c *Cond) Signal() {
	select {
	case c.signal <- lang.Placeholder:
	default:
	}
}

// NewCond 返回一个 Cond。
func NewCond() *Cond {
	return &Cond{
		signal: make(chan lang.PlaceholderType),
	}
}
