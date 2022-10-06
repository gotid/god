package syncx

import (
	"runtime"
	"sync/atomic"
)

// SpinLock 自旋锁
type SpinLock struct {
	lock uint32
}

func (l *SpinLock) Lock() {
	for !l.TryLock() {
		//暂停当前goroutine，让其他goroutine先运算
		runtime.Gosched()
	}
}

// TryLock 尝试对自旋锁进行上锁。
func (l *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapUint32(&l.lock, 0, 1)
}

func (l *SpinLock) Unlock() {
	atomic.SwapUint32(&l.lock, 0)
}
