package syncx

import "sync/atomic"

// AtomicBool bool类型 原子类
type AtomicBool uint32

func NewAtomicBool() *AtomicBool {
	return new(AtomicBool)
}

func ForAtomicBool(val bool) *AtomicBool {
	b := NewAtomicBool()
	b.Set(val)
	return b
}

func (b *AtomicBool) Set(val bool) {
	if val {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}

func (b *AtomicBool) True() bool {
	return atomic.LoadUint32((*uint32)(b)) == 1
}

func (b *AtomicBool) CompareAndSwap(old, val bool) bool {
	var ov, nv uint32
	if old {
		ov = 1
	}
	if val {
		nv = 1
	}
	return atomic.CompareAndSwapUint32((*uint32)(b), ov, nv)
}
