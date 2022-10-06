package executors

import (
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
	"time"
)

// LessExecutor 是一个指定时间间隔内只执行一次的执行人。
type LessExecutor struct {
	threshold time.Duration
	lastTime  *syncx.AtomicDuration
}

// NewLessExecutor 返回一个以给定时间间隔为阈值的 LessExecutor。
func NewLessExecutor(threshold time.Duration) *LessExecutor {
	return &LessExecutor{
		threshold: threshold,
		lastTime:  syncx.NewAtomicDuration(),
	}
}

// DoOrDiscard 执行或放弃该任务取决于该时间间隔内是否执行了其他任务。
func (le *LessExecutor) DoOrDiscard(execute func()) bool {
	now := timex.Now()
	lastTime := le.lastTime.Load()
	if lastTime == 0 || lastTime+le.threshold < now {
		le.lastTime.Set(now)
		execute()
		return true
	}

	return false
}
