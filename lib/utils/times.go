package utils

import (
	"fmt"
	"time"

	"git.zc0901.com/go/god/lib/timex"
)

// ElapsedTimer 是一个跟踪耗时的计时器。
type ElapsedTimer struct {
	start time.Duration
}

// NewElapsedTimer 返回一个耗时跟踪器。
func NewElapsedTimer() *ElapsedTimer {
	return &ElapsedTimer{
		start: timex.Now(),
	}
}

// Duration 返回消耗时长。
func (t *ElapsedTimer) Duration() time.Duration {
	return timex.Since(t.start)
}

// Elapsed 返回消耗时长的字符串表达形式。
func (t *ElapsedTimer) Elapsed() string {
	return timex.Since(t.start).String()
}

// ElapsedMs 返回消耗时长的毫秒字符串。
func (t *ElapsedTimer) ElapsedMs() string {
	return fmt.Sprintf("%.1fms", float32(timex.Since(t.start))/float32(time.Millisecond))
}

// CurrentMicros 返回当前微秒。
func CurrentMicros() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

// CurrentMillis 返回当前毫秒。
func CurrentMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
