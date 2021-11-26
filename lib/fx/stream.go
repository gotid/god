package fx

import (
	"sort"
	"sync"

	"git.zc0901.com/go/god/lib/collection"

	"git.zc0901.com/go/god/lib/lang"
	"git.zc0901.com/go/god/lib/threading"
)

const (
	defaultWorkers = 16
	minWorkers     = 1
)

type (
	rxOption struct {
		unlimitedWorkers bool
		workers          int
	}

	// GenerateFunc 定义流生成函数。
	GenerateFunc func(source chan<- interface{})
	// KeyFunc 定义键生成函数。
	KeyFunc func(item interface{}) interface{}
	// ForAllFunc 定义处理流中所有元素的函数。
	ForAllFunc func(pipe <-chan interface{})
	// ForEachFunc 定义处理流中每个元素的函数。
	ForEachFunc func(item interface{})
	// WalkFunc 定义遍历流中所有元素的方法。
	WalkFunc func(item interface{}, pipe chan<- interface{})
	// ParallelFunc 定义并行处理流中所有元素的函数。
	ParallelFunc func(item interface{})
	// FilterFunc 定义过滤流中元素的函数。
	FilterFunc func(item interface{}) bool
	// MapFunc 定义将流中每个元素映射为另外对象的函数。
	MapFunc func(item interface{}) interface{}
	// ReduceFunc 定义减少流中元素的函数。
	ReduceFunc func(pipe <-chan interface{}) (interface{}, error)
	// LessFunc 定义比较流中元素的函数。
	LessFunc func(a interface{}, b interface{}) bool
	// Option 定义一个自定义流的函数。
	Option func(opts *rxOption)

	// Stream 是一个可用于流处理的结构体。
	Stream struct {
		source <-chan interface{}
	}
)

// From 从生成函数生成一个流。
func From(generate GenerateFunc) Stream {
	source := make(chan interface{})

	threading.GoSafe(func() {
		defer close(source)
		generate(source)
	})

	return Range(source)
}

// Concat 返回一个合并后的流。
func Concat(s Stream, others ...Stream) Stream {
	return s.Concat(others...)
}

// Just 转换任意项为数据流 Stream
func Just(items ...interface{}) Stream {
	source := make(chan interface{}, len(items))
	for _, item := range items {
		source <- item
	}
	close(source)

	return Range(source)
}

// Range 将指定通道转换为 Stream 流
func Range(source <-chan interface{}) Stream {
	return Stream{source: source}
}

// AllMatch 判断流中是否所有元素都匹配指定断言。
// 有一个不满足就返回 false。
// 空流默认为满足，返回 true。
func (s Stream) AllMatch(predicate func(item interface{}) bool) bool {
	for item := range s.source {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// AnyMatch 判断流中是否有任一元素匹配指定断言。
// 有一个满足就返回 true。
// 空流默认为不满足，返回 false。
func (s Stream) AnyMatch(predicate func(item interface{}) bool) bool {
	for item := range s.source {
		if predicate(item) {
			return true
		}
	}
	return false
}

// Buffer 缓冲流中元素到一个大小为 n 的通道。
func (s Stream) Buffer(n int) Stream {
	if n < 0 {
		n = 0
	}

	source := make(chan interface{}, n)
	go func() {
		for item := range s.source {
			source <- item
		}
		close(source)
	}()

	return Range(source)
}

// Concat 返回一个连接其他流的流。
func (s Stream) Concat(others ...Stream) Stream {
	source := make(chan interface{})

	go func() {
		group := threading.NewRoutineGroup()
		group.Run(func() {
			for item := range s.source {
				source <- item
			}
		})

		for _, other := range others {
			other := other
			group.Run(func() {
				for item := range other.source {
					source <- item
				}
			})
		}

		group.Wait()
		close(source)
	}()

	return Range(source)
}

// Count 计算流中的元素个数。
func (s Stream) Count() (count int) {
	for range s.source {
		count++
	}
	return
}

// Distinct 基于指定的键函数移除重复元素。
func (s Stream) Distinct(fn KeyFunc) Stream {
	source := make(chan interface{})

	threading.GoSafe(func() {
		defer close(source)

		keys := make(map[interface{}]lang.PlaceholderType)
		for item := range s.source {
			key := fn(item)
			if _, ok := keys[key]; !ok {
				source <- item
				keys[key] = lang.Placeholder
			}
		}
	})

	return Range(source)
}

// Done 等待所有上游操作完成。
func (s Stream) Done() {
	for range s.source {
	}
}

// Filter 通过指定过滤函数过滤并返回新流。
func (s Stream) Filter(fn FilterFunc, opts ...Option) Stream {
	return s.Walk(func(item interface{}, pipe chan<- interface{}) {
		if fn(item) {
			pipe <- item
		}
	}, opts...)
}

// ForAll 处理当前流的所有元素。
func (s Stream) ForAll(fn ForAllFunc) {
	fn(s.source)
}

// ForEach 处理当前流的每个元素。
func (s Stream) ForEach(fn ForEachFunc) {
	for item := range s.source {
		fn(item)
	}
}

// Group 基于指定的键函数进行分组。
func (s Stream) Group(fn KeyFunc) Stream {
	groups := make(map[interface{}][]interface{})
	for item := range s.source {
		key := fn(item)
		groups[key] = append(groups[key], item)
	}

	source := make(chan interface{})
	go func() {
		for _, item := range groups {
			source <- item
		}
		close(source)
	}()

	return Range(source)
}

// Head 返回流中的前N个元素。
func (s Stream) Head(n int64) Stream {
	if n < 1 {
		panic("n 必须大于 0")
	}

	source := make(chan interface{})

	go func() {
		for item := range s.source {
			n--
			if n >= 0 {
				source <- item
			}
			if n == 0 {
				close(source)
			}
		}
		if n > 0 {
			close(source)
		}
	}()

	return Range(source)
}

func (s Stream) Tail(n int64) Stream {
	if n < 1 {
		panic("n 应大于 0")
	}

	source := make(chan interface{})

	go func() {
		ring := collection.NewRing(int(n))
		for item := range s.source {
			ring.Add(item)
		}
		for _, item := range ring.Take() {
			source <- item
		}
		close(source)
	}()

	return Range(source)
}

// Map 将流中每个元素映射为其他对象。
func (s Stream) Map(fn MapFunc, opts ...Option) Stream {
	return s.Walk(func(item interface{}, pipe chan<- interface{}) {
		pipe <- fn(item)
	}, opts...)
}

// Merge 展平流中所有元素为一个切片并生成新流。
// 已弃用: 使用 Flatten。
func (s Stream) Merge() Stream {
	return s.Flatten()
}

// Flatten 展平流中所有元素为一个切片并生成新流。
func (s Stream) Flatten() Stream {
	var items []interface{}
	for item := range s.source {
		items = append(items, item)
	}

	source := make(chan interface{})
	source <- items
	close(source)

	return Range(source)
}

// Parallel 使用指定协程并行应用指定的并行函数至流中各项。
func (s Stream) Parallel(fn ParallelFunc, opts ...Option) {
	s.Walk(func(item interface{}, pipe chan<- interface{}) {
		fn(item)
	}, opts...).Done()
}

// Reduce 使用指定减少函数处理流的底层通道。
func (s Stream) Reduce(fn ReduceFunc) (interface{}, error) {
	return fn(s.source)
}

// Reverse 反转流中元素。
func (s Stream) Reverse() Stream {
	var items []interface{}
	for item := range s.source {
		items = append(items, item)
	}
	// 反转，官方方法
	for i := len(items)/2 - 1; i >= 0; i-- {
		opp := len(items) - 1 - i
		items[i], items[opp] = items[opp], items[i]
	}

	return Just(items...)
}

// Skip 返回跳过前N个元素的新流。
func (s Stream) Skip(n int64) Stream {
	if n < 0 {
		panic("n 不能为负数")
	}
	if n == 0 {
		return s
	}

	source := make(chan interface{})

	go func() {
		for item := range s.source {
			n--
			if n >= 0 {
				continue
			} else {
				source <- item
			}
		}
		close(source)
	}()

	return Range(source)
}

// Sort 使用指定排序函数对源进行排序。
func (s Stream) Sort(less LessFunc) Stream {
	var items []interface{}
	for item := range s.source {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return less(items[i], items[j])
	})

	return Just(items...)
}

// Split 将流分割为元素个数不大于N的块，尾块个数可能小于N。
func (s Stream) Split(n int) Stream {
	if n < 1 {
		panic("n 应大于 0")
	}

	source := make(chan interface{})

	go func() {
		var chunk []interface{}
		for item := range s.source {
			chunk = append(chunk, item)
			if len(chunk) == n {
				source <- chunk
				chunk = nil
			}
		}
		if chunk != nil {
			source <- chunk
		}
		close(source)
	}()

	return Range(source)
}

// Walk 让调用者处理每个项，可以处理后续流。
func (s Stream) Walk(fn WalkFunc, opts ...Option) Stream {
	option := buildOptions(opts...)
	if option.unlimitedWorkers {
		return s.walkUnlimited(fn, option)
	}
	return s.walkLimited(fn, option)
}

func (s Stream) walkUnlimited(fn WalkFunc, option *rxOption) Stream {
	pipe := make(chan interface{}, defaultWorkers)

	go func() {
		var wg sync.WaitGroup

		for {
			item, ok := <-s.source
			if !ok {
				break
			}

			wg.Add(1)
			threading.GoSafe(func() {
				defer func() {
					wg.Done()
					fn(item, pipe)
				}()
			})
		}

		wg.Wait()
		close(pipe)
	}()

	return Range(pipe)
}

func (s Stream) walkLimited(fn WalkFunc, option *rxOption) Stream {
	pipe := make(chan interface{}, option.workers)

	go func() {
		var wg sync.WaitGroup
		pool := make(chan lang.PlaceholderType, option.workers)

		for {
			pool <- lang.Placeholder
			item, ok := <-s.source
			if !ok {
				<-pool
				break
			}

			wg.Add(1)
			threading.GoSafe(func() {
				defer func() {
					wg.Done()
					<-pool
				}()

				fn(item, pipe)
			})
		}

		wg.Wait()
		close(pipe)
	}()

	return Range(pipe)
}

// UnlimitedWorkers 允许调用方法使用与任务数量相同的协程。
func UnlimitedWorkers() Option {
	return func(opts *rxOption) {
		opts.unlimitedWorkers = true
	}
}

// WithWorkers 返回一个自定义并发数的函数。
func WithWorkers(workers int) Option {
	return func(opts *rxOption) {
		if workers < minWorkers {
			opts.workers = minWorkers
		} else {
			opts.workers = workers
		}
	}
}

func buildOptions(opts ...Option) *rxOption {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	return options
}

func newOptions() *rxOption {
	return &rxOption{
		workers: defaultWorkers,
	}
}
