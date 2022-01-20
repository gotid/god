package collection

import (
	"fmt"
	"sync"
	"time"

	"github.com/gotid/god/lib/timex"
)

type (
	// RollingWindow 滚动窗口计数器，维护窗口内每个桶的请求个数和次数
	RollingWindow struct {
		size                int           // 滚动窗口被分为的存储桶个数
		duration            time.Duration // 滚动窗口的持续时间段，通常为10-15秒
		win                 *window       // 内部封装的滚动窗口，维护窗口内桶的请求个数和次数
		offset              int
		ignoreCurrentBucket bool // 是否忽略当前桶
		lastTime            time.Duration
		lock                sync.RWMutex
	}

	Option func(w *RollingWindow)
)

// NewRollingWindow 新建滚窗计数器
//
// size 桶个数
//
// duration 单桶耗时
func NewRollingWindow(size int, duration time.Duration, opts ...Option) *RollingWindow {
	if size < 1 {
		panic("size 必须大于0")
	}
	w := &RollingWindow{
		size:     size,
		duration: duration,
		win:      newWindow(size),
		lastTime: timex.Now(),
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (rw *RollingWindow) Win() *window {
	return rw.win
}

func (rw *RollingWindow) Add(n float64) {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	rw.updateOffset()
	rw.win.add(rw.offset, n)
}

// Reduce 归并符合条件的桶内计数
func (rw *RollingWindow) Reduce(fn func(b *Bucket)) {
	rw.lock.RLock()
	defer rw.lock.RUnlock()

	var nDiff int
	span := rw.span()
	if span == 0 && rw.ignoreCurrentBucket {
		nDiff = rw.size - 1
	} else {
		nDiff = rw.size - span
	}
	if nDiff > 0 {
		startOffset := (rw.offset + span + 1) % rw.size
		rw.win.reduceBuckets(startOffset, nDiff, fn)
	}
}

// 更新滚动窗口偏移量（重置过期桶计数、设置当前桶offset和最新时间）
func (rw *RollingWindow) updateOffset() {
	span := rw.span() // 获取偏移跨度
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
	rw.lastTime = now - (now-rw.lastTime)%rw.duration
}

// // 获取滚动窗口的偏移跨度（根据时间差计算）
func (rw *RollingWindow) span() int {
	offset := int(timex.Since(rw.lastTime) / rw.duration)
	if 0 <= offset && offset < rw.size {
		return offset
	} else {
		return rw.size
	}
}

// IgnoreCurrentBucket 忽略当前桶
func IgnoreCurrentBucket() Option {
	return func(w *RollingWindow) {
		w.ignoreCurrentBucket = true
	}
}

// --------------- 内部滚动窗口 --------------- //
// window 内部封装的滚动窗口，维护窗口内桶的请求个数和次数
type window struct {
	size    int       // 滚动窗口被分为的存储桶个数
	buckets []*Bucket // 存储桶切片
}

// newWindow 新建内部滚动窗口
// size 滚动窗口被分为的存储桶个数
func newWindow(size int) *window {
	buckets := make([]*Bucket, size)
	for i := 0; i < size; i++ {
		buckets[i] = new(Bucket)
	}

	return &window{
		size:    size,
		buckets: buckets,
	}
}

func (w *window) Buckets() []*Bucket {
	return w.buckets
}

// 向滚动窗口指定的桶内增加 n 个请求计数
func (w *window) add(offset int, n float64) {
	// % 取模获取滚动窗口
	w.buckets[offset%w.size].add(n)
}

// 重置滚动窗口指定桶的计数器
func (w *window) resetBucket(offset int) {
	w.buckets[offset%w.size].reset()
}

// 对起始桶之后的 n 个桶进行指定操作
func (w *window) reduceBuckets(start, count int, fn func(b *Bucket)) {
	for i := 0; i < count; i++ {
		fn(w.buckets[(start+i)%w.size])
	}
}

// Bucket --------------- 滚动窗口使用的请求桶 --------------- //
// Bucket 存储桶是滚动窗口给定时间段内的请求集合
type Bucket struct {
	Requests int64   // 桶内请求次数
	Accepts  float64 // 桶内接受次数
}

// add 向当前桶增加 n 个请求计数
func (b *Bucket) add(n float64) {
	b.Requests++   // 不论是否接受，请求一次算一次
	b.Accepts += n // 大于0，接受了，才累加
}

// reset 归零桶内的请求次数和接受次数
func (b *Bucket) reset() {
	b.Requests = 0
	b.Accepts = 0
}

func (b *Bucket) String() string {
	return fmt.Sprintf("Requests: %.0d, Accepts: %.0f\n", b.Requests, b.Accepts)
}
