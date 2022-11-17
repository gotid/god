package collection

import (
	"github.com/gotid/god/lib/timex"
	"sync"
	"time"
)

type (
	// RollingWindow 滚动窗口，计算时间间隔内桶的事件数量。
	RollingWindow struct {
		lock          sync.RWMutex
		size          int           // 桶数
		win           *window       // 窗口
		interval      time.Duration // 窗口时长
		offset        int           // 当前桶的偏移量
		ignoreCurrent bool          // 是否忽略当前桶
		lastTime      time.Duration // 最后一个桶的开始时间
	}

	// RollingWindowOption 自定义 RollingWindow 的方法。
	RollingWindowOption func(rw *RollingWindow)
)

// NewRollingWindow 返回一个指定桶数、桶时长和自定义选项的 RollingWindow。
func NewRollingWindow(size int, interval time.Duration, opts ...RollingWindowOption) *RollingWindow {
	if size < 1 {
		panic("size 必须大于 0")
	}

	w := &RollingWindow{
		size:     size,
		win:      newWindow(size),
		interval: interval,
		lastTime: timex.Now(),
	}
	for _, opt := range opts {
		opt(w)
	}

	return w
}

// Add 添加 v 到当前桶。
func (rw *RollingWindow) Add(v float64) {
	rw.lock.Lock()
	defer rw.lock.Unlock()

	rw.updateOffset()
	rw.win.add(rw.offset, v)
}

// Reduce 使用 fn 聚合 Bucket，如果设置了 ignoreCurrent 则忽略当前桶。
func (rw *RollingWindow) Reduce(fn func(b *Bucket)) {
	rw.lock.Lock()
	defer rw.lock.Unlock()

	var diff int
	span := rw.span()
	//	忽略当前桶，因为局部数据
	if span == 0 && rw.ignoreCurrent {
		diff = rw.size - 1
	} else {
		diff = rw.size - span
	}
	if diff > 0 {
		offset := (rw.offset + span + 1) % rw.size
		rw.win.reduce(offset, diff, fn)
	}
}

func (rw *RollingWindow) updateOffset() {
	span := rw.span()
	if span <= 0 {
		return
	}

	offset := rw.offset
	// 重置过期桶
	for i := 0; i < span; i++ {
		rw.win.resetBucket((offset + i + 1) % rw.size)
	}

	rw.offset = (offset + span) % rw.size
	now := timex.Now()
	//对齐间隔时间边界
	rw.lastTime = now - (now-rw.lastTime)%rw.interval
}

func (rw *RollingWindow) span() int {
	offset := int(timex.Since(rw.lastTime) / rw.interval)
	if 0 <= offset && offset < rw.size {
		return offset
	}

	return rw.size
}

// Bucket 定义了保存总数和次数的桶。
type Bucket struct {
	Sum   float64 // 总数
	Count int64   // 次数
}

func (b *Bucket) add(v float64) {
	b.Sum += v
	b.Count++
}

func (b *Bucket) reset() {
	b.Sum = 0
	b.Count = 0
}

type window struct {
	buckets []*Bucket // 桶切片
	size    int       // 桶数
}

func newWindow(size int) *window {
	buckets := make([]*Bucket, size)
	for i := 0; i < size; i++ {
		buckets[i] = new(Bucket)
	}
	return &window{
		buckets: buckets,
		size:    size,
	}
}
func (w *window) add(offset int, v float64) {
	w.buckets[offset%w.size].add(v)
}

func (w *window) reduce(start, count int, fn func(b *Bucket)) {
	for i := 0; i < count; i++ {
		fn(w.buckets[(start+i)%w.size])
	}
}

func (w *window) resetBucket(offset int) {
	w.buckets[offset%w.size].reset()
}

// IgnoreCurrentBucket 让 window reduce 时忽略当前 bucket。
func IgnoreCurrentBucket() RollingWindowOption {
	return func(rw *RollingWindow) {
		rw.ignoreCurrent = true
	}
}
