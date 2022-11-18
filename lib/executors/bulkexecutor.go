package executors

import "time"

const defaultBulkTasks = 1000

type (
	// BulkOption 自定义 BulkExecutor 的方法。
	BulkOption func(options *bulkOptions)

	// BulkExecutor 可执行以下任务：
	// 1. 已达到给定任务的大小
	// 2. 已达到刷新间隔时间
	BulkExecutor struct {
		executor  *PeriodicalExecutor
		container *bulkContainer
	}

	bulkOptions struct {
		cachedTasks   int
		flushInterval time.Duration
	}
)

// NewBulkExecutor 返回一个批量执行器 BulkExecutor。
func NewBulkExecutor(execute Execute, opts ...BulkOption) *BulkExecutor {
	options := newBulkOptions()
	for _, opt := range opts {
		opt(&options)
	}

	container := &bulkContainer{
		execute:  execute,
		maxTasks: options.cachedTasks,
	}
	executor := &BulkExecutor{
		executor:  NewPeriodicalExecutor(options.flushInterval, container),
		container: container,
	}

	return executor
}

// Add 添加任务到 be。
func (be *BulkExecutor) Add(task any) error {
	be.executor.Add(task)
	return nil
}

// Flush 强制刷新并执行任务。
func (be *BulkExecutor) Flush() {
	be.executor.Flush()
}

// Wait 等待任务执行完毕。
func (be *BulkExecutor) Wait() {
	be.executor.Wait()
}

// WithBulkTasks 自定义一批任务的数量。
func WithBulkTasks(tasks int) BulkOption {
	return func(options *bulkOptions) {
		options.cachedTasks = tasks
	}
}

// WithBulkInterval 自定批量任务的刷新时间间隔。
func WithBulkInterval(interval time.Duration) BulkOption {
	return func(options *bulkOptions) {
		options.flushInterval = interval
	}
}

func newBulkOptions() bulkOptions {
	return bulkOptions{
		cachedTasks:   defaultBulkTasks,
		flushInterval: defaultFlushInterval,
	}
}

type bulkContainer struct {
	tasks    []any
	execute  Execute
	maxTasks int
}

func (bc *bulkContainer) AddTask(task any) bool {
	bc.tasks = append(bc.tasks, task)
	return len(bc.tasks) >= bc.maxTasks
}

func (bc *bulkContainer) Execute(tasks any) {
	vs := tasks.([]any)
	bc.execute(vs)
}

func (bc *bulkContainer) RemoveAll() any {
	tasks := bc.tasks
	bc.tasks = nil
	return tasks
}
