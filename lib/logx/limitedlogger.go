package logx

import (
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
)

// 指定时间内仅写一次的日志写入器
type limitedLogger struct {
	duration  time.Duration
	lastTime  *syncx.AtomicDuration
	discarded uint32
}

func newLimitedLogger(milliseconds int) *limitedLogger {
	return &limitedLogger{
		duration: time.Duration(milliseconds) * time.Millisecond,
		lastTime: syncx.NewAtomicDuration(),
	}
}

func (le *limitedLogger) logOrDiscard(fn func()) {
	if le == nil || le.duration <= 0 {
		fn()
		return
	}

	now := timex.Now()
	if now-le.lastTime.Load() <= le.duration {
		atomic.AddUint32(&le.discarded, 1)
	} else {
		le.lastTime.Set(now)
		discarded := atomic.SwapUint32(&le.discarded, 0)
		if discarded > 0 {
			Errorf("放弃 %d 个错误信息", discarded)
		}

		fn()
	}
}
