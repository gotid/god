package mr

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/gotid/god/lib/errorx"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/threading"
)

const (
	defaultWorkers = 16
	minWorkers     = 1
)

var (
	ErrCancelWithNil  = errors.New("mapreduce 已取消")
	ErrReduceNoOutput = errors.New("reduce 未写值")
)

type (
	GenerateFunc    func(source chan<- interface{})
	MapFunc         func(item interface{}, writer Writer)
	VoidMapFunc     func(item interface{})
	MapperFunc      func(item interface{}, writer Writer, cancel func(error))
	ReducerFunc     func(pipe <-chan interface{}, writer Writer, cancel func(error))
	VoidReducerFunc func(pipe <-chan interface{}, cancel func(error))
	Option          func(opts *mapReduceOptions)

	mapReduceOptions struct {
		ctx     context.Context
		workers int
	}

	Writer interface {
		Write(v interface{})
	}
)

// Finish 并行执行，有错停止。
func Finish(fns ...func() error) error {
	if len(fns) == 0 {
		return nil
	}

	return MapReduceVoid(func(source chan<- interface{}) {
		for _, fn := range fns {
			source <- fn
		}
	}, func(item interface{}, writer Writer, cancel func(error)) {
		fn := item.(func() error)
		if err := fn(); err != nil {
			cancel(err)
		}
	}, func(pipe <-chan interface{}, cancel func(error)) {
		drain(pipe)
	}, WithWorkers(len(fns)))
}

// FinishVoid 并发执行，有错继续。
func FinishVoid(fns ...func()) {
	if len(fns) == 0 {
		return
	}

	MapVoid(func(source chan<- interface{}) {
		for _, fn := range fns {
			source <- fn
		}
	}, func(item interface{}) {
		fn := item.(func())
		fn()
	}, WithWorkers(len(fns)))
}

// Map 并行映射所有元素，返回结果通道。
// opts 可选参数，目前可指定协程数量。
func Map(generate GenerateFunc, mapper MapFunc, opts ...Option) chan interface{} {
	options := buildOptions(opts...)
	source := buildSource(generate)
	collector := make(chan interface{}, options.workers)
	done := syncx.NewDoneChan()

	go executeMappers(options.ctx, mapper, source, collector, done.Done(), options.workers)

	return collector
}

// MapReduce 并行映射且聚合元素。返回结果及错误。
func MapReduce(generate GenerateFunc, mapper MapperFunc, reducer ReducerFunc, opts ...Option) (interface{}, error) {
	source := buildSource(generate)
	return MapReduceWithSource(source, mapper, reducer, opts...)
}

// MapReduceWithSource 支持传入数据源，返回聚合后的元素。
func MapReduceWithSource(source <-chan interface{}, mapper MapperFunc, reducer ReducerFunc,
	opts ...Option) (interface{}, error) {
	options := buildOptions(opts...)
	output := make(chan interface{})
	collector := make(chan interface{}, options.workers)
	done := syncx.NewDoneChan()
	writer := newGuardedWriter(options.ctx, output, done.Done())
	var closeOnce sync.Once
	var retErr errorx.AtomicError
	finish := func() {
		closeOnce.Do(func() {
			done.Close()
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

	go func() {
		defer func() {
			drain(collector)

			if r := recover(); r != nil {
				cancel(fmt.Errorf("%v", r))
			} else {
				finish()
			}
		}()
		reducer(collector, writer, cancel)
	}()

	go executeMappers(options.ctx, func(item interface{}, w Writer) {
		mapper(item, w, cancel)
	}, source, collector, done.Done(), options.workers)

	value, ok := <-output
	if err := retErr.Load(); err != nil {
		return nil, err
	} else if ok {
		return value, nil
	} else {
		return nil, ErrReduceNoOutput
	}
}

// MapReduceVoid 并行且聚合元素。无返回值，不关心错误。
func MapReduceVoid(generator GenerateFunc, mapper MapperFunc, reducer VoidReducerFunc, opts ...Option) error {
	_, err := MapReduce(generator, mapper, func(input <-chan interface{}, writer Writer, cancel func(error)) {
		reducer(input, cancel)
		drain(input)
		// We need to write a placeholder to let MapReduce to continue on reducer done,
		// otherwise, all goroutines are waiting. The placeholder will be discarded by MapReduce.
		writer.Write(lang.Placeholder)
	}, opts...)
	return err
}

// MapVoid 并行映射。无返回值，不关心错误。
func MapVoid(generate GenerateFunc, mapper VoidMapFunc, opts ...Option) {
	drain(Map(generate, func(item interface{}, writer Writer) {
		mapper(item)
	}, opts...))
}

// WithContext 自定义MapReduce的上下文。
func WithContext(ctx context.Context) Option {
	return func(opts *mapReduceOptions) {
		opts.ctx = ctx
	}
}

// WithWorkers 自定义MapReduce的worker数量。
func WithWorkers(workers int) Option {
	return func(opts *mapReduceOptions) {
		if workers < minWorkers {
			opts.workers = minWorkers
		} else {
			opts.workers = workers
		}
	}
}

func buildOptions(opts ...Option) *mapReduceOptions {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	return options
}

func buildSource(generate GenerateFunc) chan interface{} {
	source := make(chan interface{})
	threading.GoSafe(func() {
		defer close(source)
		generate(source)
	})

	return source
}

// drain drains the channel.
func drain(channel <-chan interface{}) {
	// drain the channel
	for range channel {
	}
}

func executeMappers(ctx context.Context, mapper MapFunc, input <-chan interface{}, collector chan<- interface{},
	done <-chan lang.PlaceholderType, workers int) {
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		close(collector)
	}()

	pool := make(chan lang.PlaceholderType, workers)
	writer := newGuardedWriter(ctx, collector, done)
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case pool <- lang.Placeholder:
			item, ok := <-input
			if !ok {
				<-pool
				return
			}

			wg.Add(1)
			// better to safely run caller defined method
			threading.GoSafe(func() {
				defer func() {
					wg.Done()
					<-pool
				}()

				mapper(item, writer)
			})
		}
	}
}

func newOptions() *mapReduceOptions {
	return &mapReduceOptions{
		ctx:     context.Background(),
		workers: defaultWorkers,
	}
}

func once(fn func(error)) func(error) {
	once := new(sync.Once)
	return func(err error) {
		once.Do(func() {
			fn(err)
		})
	}
}

type guardedWriter struct {
	ctx     context.Context
	channel chan<- interface{}
	done    <-chan lang.PlaceholderType
}

func newGuardedWriter(ctx context.Context, channel chan<- interface{}, done <-chan lang.PlaceholderType) guardedWriter {
	return guardedWriter{
		ctx:     ctx,
		channel: channel,
		done:    done,
	}
}

func (gw guardedWriter) Write(v interface{}) {
	select {
	case <-gw.ctx.Done():
		return
	case <-gw.done:
		return
	default:
		gw.channel <- v
	}
}
