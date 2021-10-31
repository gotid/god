package syncx

import (
	"runtime"
	"sync/atomic"
)

// SpinLock 旋转锁
type SpinLock struct {
	lock uint32
}

// Lock 对旋转锁加锁。
func (sl *SpinLock) Lock() {
	for !sl.TryLock() {
		runtime.Gosched()
	}
}

// TryLock 尝试对旋转锁加锁。
func (sl *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapUint32(&sl.lock, 0, 1)
}

// Unlock 对旋转锁解锁。
func (sl *SpinLock) Unlock() {
	atomic.StoreUint32(&sl.lock, 0)
}
