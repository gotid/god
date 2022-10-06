package syncx

import (
	"sync/atomic"
)

// AtomicBool 是一个 atomic bool 的实现。
type AtomicBool uint32

// Set 设置 atomic bool。
func (ab *AtomicBool) Set(v bool) {
	if v {
		atomic.SwapUint32((*uint32)(ab), 1)
	} else {
		atomic.SwapUint32((*uint32)(ab), 0)
	}
}

// CompareAndSwap 比较 old 和原值，若不相等则设为新值 val 并返回真，否则返回假。
func (ab *AtomicBool) CompareAndSwap(old, val bool) bool {
	var ov, nv uint32
	if old {
		ov = 1
	}
	if val {
		nv = 1
	}
	return atomic.CompareAndSwapUint32((*uint32)(ab), ov, nv)
}

// True 判断当前 atomic bool 是否为真。
func (ab *AtomicBool) True() bool {
	return atomic.LoadUint32((*uint32)(ab)) == 1
}

// NewAtomicBool 返回一个 AtomicBool。
func NewAtomicBool() *AtomicBool {
	return new(AtomicBool)
}

// ForAtomicBool 返回具有给定 d 的 AtomicBool。
func ForAtomicBool(val bool) *AtomicBool {
	ab := NewAtomicBool()
	ab.Set(val)
	return ab
}
