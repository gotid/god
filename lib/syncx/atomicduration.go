package syncx

import (
	"sync/atomic"
	"time"
)

// AtomicDuration 是一个 atomic duration 的实现。
type AtomicDuration int64

// Set 设置当前值。
func (d *AtomicDuration) Set(v time.Duration) {
	atomic.StoreInt64((*int64)(d), int64(v))
}

// CompareAndSwap 比较 old 和原值，若不相等则设为新值 val 并返回真，否则返回假。
func (d *AtomicDuration) CompareAndSwap(old, val time.Duration) bool {
	return atomic.CompareAndSwapInt64((*int64)(d), int64(old), int64(val))
}

// Load 加载当前 atomic duration。
func (d *AtomicDuration) Load() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(d)))
}

// NewAtomicDuration 返回一个 AtomicDuration。
func NewAtomicDuration() *AtomicDuration {
	return new(AtomicDuration)
}

// ForAtomicDuration 返回具有给定 d 的 AtomicDuration。
func ForAtomicDuration(d time.Duration) *AtomicDuration {
	ad := NewAtomicDuration()
	ad.Set(d)
	return ad
}
