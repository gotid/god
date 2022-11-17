package syncx

import (
	"runtime"
	"sync/atomic"
)

// SpinLock 用作快速执行的锁。
type SpinLock struct {
	lock uint32
}

// Lock 对 SpinLock 加锁。
func (l *SpinLock) Lock() {
	for !l.TryLock() {
		//暂停当前goroutine，让其他goroutine先运算
		runtime.Gosched()
	}
}

// TryLock 尝试对 SpinLock 加锁。
func (l *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapUint32(&l.lock, 0, 1)
}

// Unlock 对 SpinLock 解锁。
func (l *SpinLock) Unlock() {
	atomic.SwapUint32(&l.lock, 0)
}
