package syncx

import (
	"sync/atomic"
	"time"
)

// AtomicDuration Duration类型 原子类
type AtomicDuration int64

func NewAtomicDuration() *AtomicDuration {
	return new(AtomicDuration)
}

func ForAtomicDuration(duration time.Duration) *AtomicDuration {
	d := NewAtomicDuration()
	d.Set(duration)
	return d
}

func (ad *AtomicDuration) Set(val time.Duration) {
	atomic.StoreInt64((*int64)(ad), int64(val))
}

func (ad *AtomicDuration) Load() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(ad)))
}

func (ad *AtomicDuration) CompareAndSwap(old, new time.Duration) bool {
	return atomic.CompareAndSwapInt64((*int64)(ad), int64(old), int64(new))
}
