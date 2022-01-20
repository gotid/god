package threading

import (
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/rescue"
)

// TaskRunner 用于控制协程数的并发任务执行者。
type TaskRunner struct {
	limitChan chan lang.PlaceholderType
}

// NewTaskRunner 返回一个可控制协程并发数的任务执行者。
func NewTaskRunner(concurrency int) *TaskRunner {
	return &TaskRunner{
		limitChan: make(chan lang.PlaceholderType, concurrency),
	}
}

// Schedule 安排任务在并发控制下运行。
func (r *TaskRunner) Schedule(task func()) {
	r.limitChan <- lang.Placeholder

	go func() {
		defer rescue.Recover(func() {
			<-r.limitChan
		})

		task()
	}()
}
