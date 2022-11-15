package fx

import (
	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/threading"
	"sort"
	"sync"
)

const (
	defaultWorkers = 16
	minWorkers     = 1
)

type (
	// Stream 是一个可以进行流式处理的数据流对象。
	Stream struct {
		source <-chan any
	}

	// FilterFunc 用于过滤 Stream 的方法。
	FilterFunc func(item any) bool
	// ForAllFunc 处理 Stream 所有元素的方法。
	ForAllFunc func(pipe <-chan any)
	// ForEachFunc 处理 Stream 每一个元素的方法。
	ForEachFunc func(item any)
	// GenerateFunc 生成并发送元素至 Stream 的方法。
	GenerateFunc func(source chan<- any)
	// KeyFunc 生成 Stream 中元素的键的方法。
	KeyFunc func(item any) any
	// LessFunc 比较 Stream 中元素的方法。
	LessFunc func(a, b any) bool
	// MapFunc 将 Stream 中每个元素映射为另一个对象的方法。
	MapFunc func(item any) any
	// Option 自定义 Stream 选项。
	Option func(opts *rxOptions)
	// ParallelFunc 并行处理 Stream 中每个元素的方法。
	ParallelFunc func(item any)
	// ReduceFunc 聚合 Stream 中所有元素的方法。
	ReduceFunc func(pipe <-chan any) (any, error)
	// WalkFunc 遍历 Stream 中所有元素。
	WalkFunc func(item any, pipe chan<- any)

	// Stream 选项
	rxOptions struct {
		unlimitedWorkers bool // 是否不限工作者数量
		workers          int
	}
)

// Concat 返回一个合并后的 Stream。
func Concat(stream Stream, others ...Stream) Stream {
	return stream.Concat(others...)
}

// From 从给定的 GenerateFunc 中构建一个 Stream。
func From(generate GenerateFunc) Stream {
	source := make(chan any)

	threading.GoSafe(func() {
		defer close(source)
		generate(source)
	})

	return Range(source)
}

// Just 将给定项转为 Stream。
func Just(items ...any) Stream {
	source := make(chan any, len(items))
	for _, item := range items {
		source <- item
	}
	close(source)

	return Range(source)
}

// Range 将给定的通道转换为为一个 Stream。
func Range(source <-chan any) Stream {
	return Stream{
		source: source,
	}
}

// UnlimitedWorkers 允许调用者使用与任务一样多的工作人员。
func UnlimitedWorkers() Option {
	return func(opts *rxOptions) {
		opts.unlimitedWorkers = true
	}
}

// WithWorkers 允许调用者自定义并发的工作人员数量。
func WithWorkers(workers int) Option {
	return func(opts *rxOptions) {
		if workers < minWorkers {
			opts.workers = minWorkers
		} else {
			opts.workers = workers
		}
	}
}

// AllMatch 判断 Stream 中所有元素是否都符合给定的预测。
// 如果 Stream 为空，返回 true。
func (s Stream) AllMatch(predicate func(item any) bool) bool {
	for item := range s.source {
		if !predicate(item) {
			// 确保以前的协程不被阻塞，当前函数首先返回
			go drain(s.source)

			return false
		}
	}

	return true
}

// AnyMatch 判断 Stream 中是否有任意元素满足给定的预测。
// 如果 Stream 为空，返回 false。
func (s Stream) AnyMatch(predicate func(item any) bool) bool {
	for item := range s.source {
		if predicate(item) {
			// 确保以前的协程不被阻塞，当前函数首先返回
			go drain(s.source)

			return true
		}
	}

	return false
}

// Buffer 将 Stream 中的元素缓冲到大小为 n 的新 Stream 中。
// 用于生产者和消费者的吞吐量不匹配时，进行平衡。
func (s Stream) Buffer(n int) Stream {
	if n < 0 {
		n = 0
	}

	source := make(chan any, n)
	go func() {
		for item := range s.source {
			source <- item
		}
		close(source)
	}()

	return Range(source)
}

// Concat 返回当前 Stream 与其他 Stream 合并后的新 Stream。
func (s Stream) Concat(others ...Stream) Stream {
	source := make(chan any)

	go func() {
		group := threading.NewRoutineGroup()

		// 把当前流的元素放入 source 通道
		group.Run(func() {
			for item := range s.source {
				source <- item
			}
		})

		// 把其他流的元素放入 source 通道
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

// Count 计算 Stream 中元素的数量。
func (s Stream) Count() (total int) {
	for range s.source {
		total++
	}

	return
}

// Distinct 基于给定的 KeyFunc 移除重复项。
func (s Stream) Distinct(fn KeyFunc) Stream {
	source := make(chan any)

	threading.GoSafe(func() {
		defer close(source)

		keys := make(map[any]lang.PlaceholderType)
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
	drain(s.source)
}

// Filter 使用给定的 FilterFunc 过滤元素。
func (s Stream) Filter(fn FilterFunc, opts ...Option) Stream {
	return s.Walk(func(item any, pipe chan<- any) {
		if fn(item) {
			pipe <- item
		}
	}, opts...)
}

// First 返回第一个元素。
func (s Stream) First() any {
	for item := range s.source {
		// 确保以前的协程不被阻塞，当前函数首先返回
		go drain(s.source)
		return item
	}

	return nil
}

// ForAll 处理当前 Stream 中的所有元素，不处理之后入流的元素。
func (s Stream) ForAll(fn ForAllFunc) {
	fn(s.source)
	// 避免 fn 未消费所有元素的情况下引起的协程泄露。
	go drain(s.source)
}

// ForEach 封装 Stream 中的每个元素。
func (s Stream) ForEach(fn ForEachFunc) {
	for item := range s.source {
		fn(item)
	}
}

// Group 基于给定的 KeyFunc 对 Stream 的元素进行分组。
func (s Stream) Group(fn KeyFunc) Stream {
	groups := make(map[any][]any)
	for item := range s.source {
		key := fn(item)
		groups[key] = append(groups[key], item)
	}

	source := make(chan any)
	go func() {
		for _, group := range groups {
			source <- group
		}
		close(source)
	}()

	return Range(source)
}

// Head 返回 Stream 中的前 n 个元素组成的 Stream。
func (s Stream) Head(n int64) Stream {
	if n < 1 {
		panic("n 必须大于 0")
	}

	source := make(chan any)

	go func() {
		for item := range s.source {
			n--
			if n >= 0 {
				source <- item
			}
			if n == 0 {
				// 及时需要跳过更多元素，也要让后来的方法尽快执行，故此关闭。
				close(source)

				// 为何不使用 break 跳出循坏，并排空所有元素？
				// 因为 break 会导致以前的协程永远阻塞，造成协程泄露。
				drain(s.source)
			}
		}

		// 流中源元素不足，也要让后面的方法尽快执行，故此关闭。
		if n > 0 {
			close(source)
		}
	}()

	return Range(source)
}

// Last 返回 Stream 中最后一个元素。
func (s Stream) Last() (item any) {
	for item = range s.source {
	}

	return
}

// Map 将 Stream 中的 每一个元素转换为对一个的另一个元素。
func (s Stream) Map(fn MapFunc, opts ...Option) Stream {
	return s.Walk(func(item any, pipe chan<- any) {
		pipe <- fn(item)
	}, opts...)
}

// Merge 将 Stream 中的所有元素合并到一个切片并生成一个新的 Stream。
func (s Stream) Merge() Stream {
	var items []any
	for item := range s.source {
		items = append(items, item)
	}

	source := make(chan any, 1)
	source <- items
	close(source)

	return Range(source)
}

// NoneMatch 判断是否 Stream 中所有元素都不满足给定的 predicate。
// 如果 Stream 为空，则返回 true。
func (s Stream) NoneMatch(predicate func(item any) bool) bool {
	for item := range s.source {
		if predicate(item) {
			// 确保以前的协程不被阻塞，当前函数首先返回
			go drain(s.source)
			return false
		}
	}

	return true
}

// Parallel 使用给定个数的工作者并行运行给定的 ParallelFunc。
func (s Stream) Parallel(fn ParallelFunc, opts ...Option) {
	s.Walk(func(item any, pipe chan<- any) {
		fn(item)
	}, opts...).Done()
}

// Reduce 是一个允许调用者封装底层通道的聚合方法。
func (s Stream) Reduce(fn ReduceFunc) (any, error) {
	return fn(s.source)
}

// Reverse 反转 Stream 中的元素顺序。
func (s Stream) Reverse() Stream {
	var items []any
	for item := range s.source {
		items = append(items, item)
	}

	// 反转、官方方法
	for i := len(items)/2 - 1; i >= 0; i-- {
		opp := len(items) - 1 - i
		items[i], items[opp] = items[opp], items[i]
	}

	return Just(items...)
}

// Skip 返回一个跳过 n 个元素的新 Stream。
func (s Stream) Skip(n int64) Stream {
	if n < 0 {
		panic("n 不能为负数")
	}
	if n == 0 {
		return s
	}

	source := make(chan any)

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

// Sort 对 Stream 中的元素进行排序。
func (s Stream) Sort(less LessFunc) Stream {
	var items []any
	for item := range s.source {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return less(items[i], items[j])
	})

	return Just(items...)
}

// Split 将 Stream 中的元素分成 n 块。
// 可能尾部块的元素数量少于 n。
func (s Stream) Split(n int) Stream {
	if n < 1 {
		panic("n 必须大于 0")
	}

	source := make(chan any)
	go func() {
		var chunk []any
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

// Tail 返回 Stream 中的后 n 个元素组成的 Stream。
func (s Stream) Tail(n int64) Stream {
	if n < 1 {
		panic("n 必须大于 0")
	}

	source := make(chan any)

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

// Walk 让调用者处理每个元素。调用者基于给定的元素可能会写入0个、1个或N个元素。
func (s Stream) Walk(fn WalkFunc, opts ...Option) Stream {
	option := buildOptions(opts...)
	if option.unlimitedWorkers {
		return s.walkUnlimited(fn, option)
	}

	return s.walkLimited(fn, option)
}

func (s Stream) walkLimited(fn WalkFunc, option *rxOptions) Stream {
	pipe := make(chan any, option.workers)

	go func() {
		var wg sync.WaitGroup
		pool := make(chan lang.PlaceholderType, option.workers)

		for item := range s.source {
			// 重要，用于其他协程
			val := item
			pool <- lang.Placeholder
			wg.Add(1)

			threading.GoSafe(func() {
				defer func() {
					wg.Done()
					<-pool
				}()

				fn(val, pipe)
			})
		}

		wg.Wait()
		close(pipe)
	}()

	return Range(pipe)

}

func (s Stream) walkUnlimited(fn WalkFunc, option *rxOptions) Stream {
	pipe := make(chan any, option.workers)

	go func() {
		var wg sync.WaitGroup

		for item := range s.source {
			// 重要，用于其他协程
			val := item
			wg.Add(1)

			threading.GoSafe(func() {
				defer wg.Done()
				fn(val, pipe)
			})
		}

		wg.Wait()
		close(pipe)
	}()

	return Range(pipe)
}

// 基于自定义选项返回一个 rxOptions。
func buildOptions(opts ...Option) *rxOptions {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	return options
}

// newOptions 返回一个默认的 rxOptions。
func newOptions() *rxOptions {
	return &rxOptions{
		workers: defaultWorkers,
	}
}

// drain 排干给定的通道。
func drain(channel <-chan any) {
	for range channel {
	}
}
