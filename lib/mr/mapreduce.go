package mr

import (
	"errors"
	"fmt"
	"sync"

	"git.zc0901.com/go/god/lib/errorx"
	"git.zc0901.com/go/god/lib/lang"
	"git.zc0901.com/go/god/lib/syncx"
	"git.zc0901.com/go/god/lib/threading"
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
		workers int
	}

	Writer interface {
		Write(v interface{})
	}
)

// Finish 并行执行，有错返错。
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

// FinishVoid 并行执行，不返错误。
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
func Map(generate GenerateFunc, mapper MapFunc, opts ...Option) chan interface{} {
	options := buildOptions(opts...)
	source := buildSource(generate)
	collector := make(chan interface{}, options.workers)
	done := syncx.NewDoneChan()

	go executeMappers(mapper, source, collector, done.Done(), options.workers)

	return collector
}

// MapReduce 并行映射并减少元素，返回结果及错误。
func MapReduce(generate GenerateFunc, mapper MapperFunc, reducer ReducerFunc, opts ...Option) (interface{}, error) {
	source := buildSource(generate)
	return MapReduceWithSource(source, mapper, reducer, opts...)
}

// MapReduceWithSource 映射源中的所有元素，并使用指定的reducer归纳输出的元素。
func MapReduceWithSource(source <-chan interface{}, mapper MapperFunc, reducer ReducerFunc,
	opts ...Option) (interface{}, error) {
	options := buildOptions(opts...)
	output := make(chan interface{})
	collector := make(chan interface{}, options.workers)
	done := syncx.NewDoneChan()
	writer := newGuardedWriter(output, done.Done())
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

	go executeMappers(func(item interface{}, w Writer) {
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

func MapVoid(generate GenerateFunc, mapper VoidMapFunc, opts ...Option) {
	drain(Map(generate, func(item interface{}, writer Writer) {
		mapper(item)
	}, opts...))
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

func executeMappers(mapper MapFunc, input <-chan interface{}, collector chan<- interface{},
	done <-chan lang.PlaceholderType, workers int) {
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		close(collector)
	}()

	pool := make(chan lang.PlaceholderType, workers)
	writer := newGuardedWriter(collector, done)
	for {
		select {
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
	channel chan<- interface{}
	done    <-chan lang.PlaceholderType
}

func newGuardedWriter(channel chan<- interface{}, done <-chan lang.PlaceholderType) guardedWriter {
	return guardedWriter{
		channel: channel,
		done:    done,
	}
}

func (gw guardedWriter) Write(v interface{}) {
	select {
	case <-gw.done:
		return
	default:
		gw.channel <- v
	}
}
