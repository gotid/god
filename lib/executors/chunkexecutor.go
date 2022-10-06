package executors

import "time"

const defaultChunkSize = 1024 * 1024 // 1M

type (
	// ChunkOption 自定义 ChunkExecutor 的方法。
	ChunkOption func(options *chunkOptions)

	// ChunkExecutor 可执行以下任务：
	// 1. 已达到给定的块大小
	// 2. 已达到刷新间隔时间
	ChunkExecutor struct {
		executor  *PeriodicalExecutor
		container *chunkContainer
	}

	chunkOptions struct {
		chunkSize     int
		flushInterval time.Duration
	}
)

// NewChunkExecutor 返回一个按块大小执行的执行器 ChunkExecutor。
func NewChunkExecutor(execute Execute, opts ...ChunkOption) *ChunkExecutor {
	options := newChunkOptions()
	for _, opt := range opts {
		opt(&options)
	}

	container := &chunkContainer{
		execute:      execute,
		maxChunkSize: options.chunkSize,
	}
	executor := &ChunkExecutor{
		executor:  NewPeriodicalExecutor(options.flushInterval, container),
		container: container,
	}

	return executor
}

// Add 添加任务到 be。
func (be *ChunkExecutor) Add(task interface{}, size int) error {
	be.executor.Add(chunk{
		val:  task,
		size: size,
	})
	return nil
}

// Flush 强制刷新并执行任务。
func (be *ChunkExecutor) Flush() {
	be.executor.Flush()
}

// Wait 等待任务执行完毕。
func (be *ChunkExecutor) Wait() {
	be.executor.Wait()
}

// WithChunkBytes 自定义块大小。
func WithChunkBytes(size int) ChunkOption {
	return func(options *chunkOptions) {
		options.chunkSize = size
	}
}

// WithFlushInterval 自定块执行器的刷新间隔。
func WithFlushInterval(interval time.Duration) ChunkOption {
	return func(options *chunkOptions) {
		options.flushInterval = interval
	}
}

func newChunkOptions() chunkOptions {
	return chunkOptions{
		chunkSize:     defaultChunkSize,
		flushInterval: defaultFlushInterval,
	}
}

type chunkContainer struct {
	tasks        []interface{}
	execute      Execute
	size         int
	maxChunkSize int
}

func (bc *chunkContainer) AddTask(task interface{}) bool {
	ck := task.(chunk)
	bc.tasks = append(bc.tasks, ck.val)
	bc.size += ck.size
	return bc.size >= bc.maxChunkSize
}

func (bc *chunkContainer) Execute(tasks interface{}) {
	vs := tasks.([]interface{})
	bc.execute(vs)
}

func (bc *chunkContainer) RemoveAll() interface{} {
	tasks := bc.tasks
	bc.tasks = nil
	bc.size = 0
	return tasks
}

type chunk struct {
	val  interface{}
	size int
}
