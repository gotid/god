package stat

import "time"

// Task 向 Metrics 汇报的任务。
type Task struct {
	Drop        bool
	Duration    time.Duration
	Description string
}
