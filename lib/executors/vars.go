package executors

import "time"

const defaultFlushInterval = time.Second

// Execute 定义执行任务的方法。
type Execute func(tasks []any)
