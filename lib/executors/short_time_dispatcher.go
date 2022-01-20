package executors

import (
	"time"

	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
)

// ShortTimeExecutor 短时执行器，仅在指定时间段内执行分发任务
type ShortTimeExecutor struct {
	duration time.Duration
	lastTime *syncx.AtomicDuration
}

// NewShortTimeExecutor 仅在指定时间段内执行分发任务
func NewShortTimeExecutor(duration time.Duration) *ShortTimeExecutor {
	return &ShortTimeExecutor{
		duration: duration,
		lastTime: syncx.NewAtomicDuration(),
	}
}

func (d *ShortTimeExecutor) DoOrDiscard(execute func()) bool {
	now := timex.Now()
	lastTime := d.lastTime.Load()
	if lastTime == 0 || lastTime+d.duration < now {
		d.lastTime.Set(now)
		execute()
		return true
	}

	return false
}
