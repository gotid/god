package utils

import (
	"fmt"
	"github.com/gotid/god/lib/timex"
	"time"
)

// A ElapsedTimer 是一个跟踪耗时的计时器。
type ElapsedTimer struct {
	start time.Duration
}

// NewElapsedTimer 返回一个耗时计时器 ElapsedTimer.
func NewElapsedTimer() *ElapsedTimer {
	return &ElapsedTimer{
		start: timex.Now(),
	}
}

// Duration 返回消耗时长。
func (et *ElapsedTimer) Duration() time.Duration {
	return timex.Since(et.start)
}

// Elapsed 返回耗时的字符串表示。
func (et *ElapsedTimer) Elapsed() string {
	return timex.Since(et.start).String()
}

// ElapsedMs 返回耗时的毫秒字符串表示。
func (et *ElapsedTimer) ElapsedMs() string {
	return fmt.Sprintf("%.1fms", float32(timex.Since(et.start))/float32(time.Millisecond))
}

// CurrentMicros 返回当前微秒。
func CurrentMicros() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

// CurrentMillis 返回当前毫秒。
func CurrentMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
