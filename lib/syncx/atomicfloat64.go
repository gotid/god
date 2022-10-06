package syncx

import (
	"math"
	"sync/atomic"
)

// AtomicFloat64 是一个 atomic float 64 的实现。
type AtomicFloat64 uint64

// Set 设置当前值。
func (af *AtomicFloat64) Set(val float64) {
	atomic.SwapUint64((*uint64)(af), math.Float64bits(val))
}

// Load 加载当前值。
func (af *AtomicFloat64) Load() float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)(af)))
}

// CompareAndSwap 比较 old 和原值，若不相等则设为新值 val 并返回真，否则返回假。
func (af *AtomicFloat64) CompareAndSwap(old, val float64) bool {
	return atomic.CompareAndSwapUint64((*uint64)(af), math.Float64bits(old), math.Float64bits(val))
}

// Add 累加 v 到当前值。
func (af *AtomicFloat64) Add(val float64) float64 {
	for {
		old := af.Load()
		nv := old + val
		if af.CompareAndSwap(old, nv) {
			return nv
		}
	}
}

// NewAtomicFloat64 返回一个 AtomicFloat64。
func NewAtomicFloat64() *AtomicFloat64 {
	return new(AtomicFloat64)
}

// ForAtomicFloat64 返回具有指定 val 的 AtomicFloat64。
func ForAtomicFloat64(val float64) *AtomicFloat64 {
	af := NewAtomicFloat64()
	af.Set(val)
	return af
}
