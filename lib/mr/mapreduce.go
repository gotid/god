package mr

import (
	"context"
	"errors"
	"github.com/gotid/god/lib/errorx"
	"github.com/gotid/god/lib/lang"
	"sync"
	"sync/atomic"
)

const (
	minWorkers     = 1
	defaultWorkers = 16
)

var (
	// ErrCancelWithNil 是一个 mapreduce 没有错误而进行取消的错误。
	ErrCancelWithNil = errors.New("mapreduce无错取消")
	// ErrReduceNoOutput 是一个聚合者未输出值的错误。
	ErrReduceNoOutput = errors.New("聚合者未写入输出值")
)

type (
	// ForEachFunc 处理元素，但不输出。
	ForEachFunc func(item any)
	// GenerateFunc 让调用者发送元素至数据源。
	GenerateFunc func(source chan<- any)
	// MapFunc 加工元素并将输出写入 Writer。
	MapFunc func(item any, writer Writer)
	// MapperFunc 加工元素并将输出写入 Writer。
	// 使用 cancel 函数取消处理。
	MapperFunc func(item any, writer Writer, cancel func(error))
	// ReducerFunc 聚合所有的加工输出并再输出至 Writer。
	// 使用 cancel 函数取消处理。
	ReducerFunc func(pipe <-chan any, writer Writer, cancel func(error))
	// VoidReducerFunc 聚合所有加工输出，但不再输出。
	// 使用 cancel 函数取消处理。
	VoidReducerFunc func(pipe <-chan any, cancel func(error))
	Option          func(opts *mapReduceOptions)

	mapReduceOptions struct {
		ctx     context.Context
		workers int
	}

	mapperContext struct {
		ctx       context.Context
		mapper    MapFunc
		source    <-chan any
		panicChan *onceChan
		collector chan<- any
		doneChan  <-chan lang.PlaceholderType
		workers   int
	}

	// Writer 接口包装 Write 方法。
	Writer interface {
		Write(v any)
	}
)

// Finish 并行运行 fns，遇错取消。
func Finish(fns ...func() error) error {
	if len(fns) == 0 {
		return nil
	}

	return MapReduceVoid(func(source chan<- any) {
		for _, fn := range fns {
			source <- fn
		}
	}, func(item any, writer Writer, cancel func(error)) {
		fn := item.(func() error)
		if err := fn(); err != nil {
			cancel(err)
		}
	}, func(pipe <-chan any, cancel func(error)) {
	}, WithWorkers(len(fns)))
}

// FinishVoid 并行运行 fns，忽略执行错误。
func FinishVoid(fns ...func()) {
	if len(fns) == 0 {
		return
	}

	ForEach(func(source chan<- any) {
		for _, fn := range fns {
			source <- fn
		}
	}, func(item any) {
		fn := item.(func())
		fn()
	}, WithWorkers(len(fns)))
}

// ForEach 加工所有生成的元素，但并不输出。
func ForEach(generate GenerateFunc, mapper ForEachFunc, opts ...Option) {
	options := buildOptions(opts...)
	panicChan := &onceChan{channel: make(chan interface{})}
	source := buildSource(generate, panicChan)
	collector := make(chan interface{})
	done := make(chan lang.PlaceholderType)

	go executeMappers(mapperContext{
		ctx: options.ctx,
		mapper: func(item any, _ Writer) {
			mapper(item)
		},
		source:    source,
		panicChan: panicChan,
		collector: collector,
		doneChan:  done,
		workers:   options.workers,
	})

	for {
		select {
		case v := <-panicChan.channel:
			panic(v)
		case _, ok := <-collector:
			if !ok {
				return
			}
		}
	}
}

// MapReduceVoid 加工所有生成的元素并聚合，但不输出结果。
func MapReduceVoid(generate GenerateFunc, mapper MapperFunc, reducer VoidReducerFunc, opts ...Option) error {
	_, err := MapReduce(generate, mapper, func(pipe <-chan any, writer Writer, cancel func(error)) {
		reducer(pipe, cancel)
	}, opts...)
	if errors.Is(err, ErrReduceNoOutput) {
		return nil
	}

	return err
}

// MapReduce 加工所有生成的元素，并聚合后输出。
func MapReduce(generate GenerateFunc, mapper MapperFunc, reducer ReducerFunc, opts ...Option) (any, error) {
	panicChan := &onceChan{channel: make(chan any)}
	source := buildSource(generate, panicChan)
	return mapReduceWithPanicChan(source, panicChan, mapper, reducer, opts...)
}

// MapReduceChan 加工所有给定的源数据，并聚合输出。
func MapReduceChan(source <-chan any, mapper MapperFunc, reducer ReducerFunc, opts ...Option) (any, error) {
	panicChan := &onceChan{channel: make(chan any)}
	return mapReduceWithPanicChan(source, panicChan, mapper, reducer, opts...)
}

// WithWorkers 自定义 mapreduce 的并行个数。
func WithWorkers(workers int) Option {
	return func(opts *mapReduceOptions) {
		if workers < minWorkers {
			opts.workers = minWorkers
		} else {
			opts.workers = workers
		}
	}
}

// WithContext 定义 mapreduce 的上下文。
func WithContext(ctx context.Context) Option {
	return func(opts *mapReduceOptions) {
		opts.ctx = ctx
	}
}

// 加工数据源中所有元素，并聚合后输出。
func mapReduceWithPanicChan(source <-chan any, panicChan *onceChan, mapper MapperFunc, reducer ReducerFunc, opts ...Option) (any, error) {
	options := buildOptions(opts...)

	// out 用于写入最终结果
	output := make(chan any)
	defer func() {
		// 聚合只允许写入一次，否则 panic
		for range output {
			panic("多次写入聚合器")
		}
	}()

	// collector 用于采集加工的数据，并在聚合器中消费
	collector := make(chan any, options.workers)
	// done 通道一旦关闭，所有加工者和聚合者都应停止工作
	done := make(chan lang.PlaceholderType)
	writer := newGuardedWriter(options.ctx, output, done)
	var closeOnce sync.Once
	// 使用 atomic.Value 以避免数据竞争
	var retErr errorx.AtomicError
	finish := func() {
		closeOnce.Do(func() {
			close(done)
			close(output)
		})
	}
	cancel := once(func(err error) {
		if err != nil {
			retErr.Set(err)
		} else {
			retErr.Set(ErrCancelWithNil)
		}

		drain(source)
		finish()
	})

	// 聚合数据
	go func() {
		defer func() {
			drain(collector)
			if r := recover(); r != nil {
				panicChan.write(r)
			}
			finish()
		}()

		reducer(collector, writer, cancel)
	}()

	// 加工数据
	go executeMappers(mapperContext{
		ctx: options.ctx,
		mapper: func(item any, writer Writer) {
			mapper(item, writer, cancel)
		},
		source:    source,
		panicChan: panicChan,
		collector: collector,
		doneChan:  done,
		workers:   options.workers,
	})

	select {
	case <-options.ctx.Done():
		cancel(context.DeadlineExceeded)
		return nil, context.DeadlineExceeded
	case v := <-panicChan.channel:
		// 在此排出输出通道，否则会引发 defer 中的 panic 死循环
		drain(output)
		panic(v)
	case v, ok := <-output:
		if err := retErr.Load(); err != nil {
			return nil, err
		} else if ok {
			return v, nil
		} else {
			return nil, ErrReduceNoOutput
		}
	}
}

func executeMappers(mCtx mapperContext) {
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		close(mCtx.collector)
		drain(mCtx.source)
	}()

	var failed int32
	pool := make(chan lang.PlaceholderType, mCtx.workers)
	writer := newGuardedWriter(mCtx.ctx, mCtx.collector, mCtx.doneChan)
	for atomic.LoadInt32(&failed) == 0 {
		select {
		case <-mCtx.ctx.Done():
			return
		case <-mCtx.doneChan:
			return
		case pool <- lang.Placeholder:
			item, ok := <-mCtx.source
			if !ok {
				<-pool
				return
			}

			wg.Add(1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						atomic.AddInt32(&failed, 1)
						mCtx.panicChan.write(r)
					}

					wg.Done()
					<-pool
				}()

				mCtx.mapper(item, writer)
			}()
		}
	}
}

// drain 排干给定的通道。
func drain(channel <-chan any) {
	for range channel {
	}
}

func once(fn func(err error)) func(error) {
	oc := new(sync.Once)
	return func(err error) {
		oc.Do(func() {
			fn(err)
		})
	}
}

func buildOptions(opts ...Option) *mapReduceOptions {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	return options
}

func newOptions() *mapReduceOptions {
	return &mapReduceOptions{
		ctx:     context.Background(),
		workers: defaultWorkers,
	}
}

func buildSource(generate GenerateFunc, panicChan *onceChan) chan any {
	source := make(chan any)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicChan.write(r)
			}
			close(source)
		}()

		generate(source)
	}()

	return source
}

type onceChan struct {
	channel chan any
	wrote   int32
}

func (c *onceChan) write(v any) {
	if atomic.CompareAndSwapInt32(&c.wrote, 0, 1) {
		c.channel <- v
	}
}

type guardedWriter struct {
	ctx     context.Context
	channel chan<- any
	done    <-chan lang.PlaceholderType
}

func newGuardedWriter(ctx context.Context, channel chan<- any, done <-chan lang.PlaceholderType) guardedWriter {
	return guardedWriter{
		ctx:     ctx,
		channel: channel,
		done:    done,
	}
}

func (w guardedWriter) Write(v any) {
	select {
	case <-w.ctx.Done():
		return
	case <-w.done:
		return
	default:
		w.channel <- v
	}
}
