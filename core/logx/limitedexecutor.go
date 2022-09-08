package logx

import (
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/timex"

	"github.com/gotid/god/lib/syncx"
)

type limitedExecutor struct {
	threshold time.Duration
	lastTime  *syncx.AtomicDuration
	discarded uint32
}

// 返回一个间隔时间大于给定毫秒数的受限执行器。
func newLimitedExecutor(milliseconds int) *limitedExecutor {
	return &limitedExecutor{
		threshold: time.Duration(milliseconds) * time.Millisecond,
		lastTime:  syncx.NewAtomicDuration(),
	}
}

func (le *limitedExecutor) logOrDiscard(execute func()) {
	if le == nil || le.threshold <= 0 {
		execute()
		return
	}

	now := timex.Now()
	if now-le.lastTime.Load() <= le.threshold {
		atomic.AddUint32(&le.discarded, 1)
	} else {
		le.lastTime.Set(now)
		discarded := atomic.SwapUint32(&le.discarded, 0)
		if discarded > 0 {
			Errorf("丢弃 %d 条错误消息", discarded)
		}

		execute()
	}
}
