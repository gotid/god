package threading

import (
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/rescue"
)

// TaskRunner 用于控制协程并发。
type TaskRunner struct {
	limitChan chan lang.PlaceholderType
}

// NewTaskRunner 返回一个给定并发数的 TaskRunner。
func NewTaskRunner(concurrency int) *TaskRunner {
	return &TaskRunner{
		limitChan: make(chan lang.PlaceholderType, concurrency),
	}
}

// Schedule 在并发控制下安排任务执行。
func (r *TaskRunner) Schedule(task func()) {
	r.limitChan <- lang.Placeholder

	go func() {
		defer rescue.Recover(func() {
			<-r.limitChan
		})

		task()
	}()
}
