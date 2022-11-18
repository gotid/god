package stat

import (
	"github.com/gotid/god/lib/executors"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"os"
	"sync"
	"time"
)

var (
	// 是否记录指标的统计日志
	logEnabled = syncx.ForAtomicBool(true)
	// 默认每分钟记录一次
	logInterval  = time.Minute
	writeLock    sync.Mutex
	reportWriter Writer = nil
)

type (
	// Writer 接口包装 Write 方法。
	Writer interface {
		Write(report *StatReport) error
	}

	// StatReport 是一条统计报告条目。
	StatReport struct {
		Name          string  `json:"name"`
		Timestamp     int64   `json:"tm"`
		Pid           int     `json:"pid"`
		ReqsPerSecond float32 `json:"qps"`
		Drops         int     `json:"drops"`
		Average       float32 `json:"avg"`
		Median        float32 `json:"med"`
		Top90th       float32 `json:"t90"`
		Top99th       float32 `json:"t99"`
		Top99p9th     float32 `json:"t99p9"`
	}

	// Metrics 用于记录和汇报统计报告。
	Metrics struct {
		executor  *executors.PeriodicalExecutor
		container *metricsContainer
	}

	tasksDurationPair struct {
		tasks    []Task
		duration time.Duration
		drops    int
	}
)

// DisableLog 禁用统计日志。
func DisableLog() {
	logEnabled.Set(false)
}

// SetReportWriter 设置报告编写器。
func SetReportWriter(writer Writer) {
	writeLock.Lock()
	reportWriter = writer
	writeLock.Unlock()
}

// NewMetrics 返回一个 Metrics。
func NewMetrics(name string) *Metrics {
	container := &metricsContainer{
		name: name,
		pid:  os.Getpid(),
	}

	return &Metrics{
		executor:  executors.NewPeriodicalExecutor(logInterval, container),
		container: container,
	}
}

// Add 添加任务到 m。
func (m *Metrics) Add(task Task) {
	m.executor.Add(task)
}

// AddDrop 添加一个 drop 到 m。
func (m *Metrics) AddDrop() {
	m.executor.Add(Task{
		Drop: true,
	})
}

// SetName 设置指标名称。
func (m *Metrics) SetName(name string) {
	m.executor.Sync(func() {
		m.container.name = name
	})
}

// 指标任务容器
type metricsContainer struct {
	name     string
	pid      int
	tasks    []Task
	duration time.Duration
	drops    int
}

func (c *metricsContainer) AddTask(v any) bool {
	if task, ok := v.(Task); ok {
		if task.Drop {
			c.drops++
		} else {
			c.tasks = append(c.tasks, task)
			c.duration += task.Duration
		}
	}

	return false
}

func (c *metricsContainer) Execute(v any) {
	pair := v.(tasksDurationPair)
	tasks := pair.tasks
	duration := pair.duration
	drops := pair.drops
	size := len(tasks)
	report := &StatReport{
		Name:          c.name,
		Timestamp:     time.Now().Unix(),
		Pid:           c.pid,
		ReqsPerSecond: float32(size) / float32(logInterval/time.Second),
		Drops:         drops,
	}

	if size > 0 {
		report.Average = float32(duration/time.Millisecond) / float32(size)

		fiftyPercent := size >> 1
		if fiftyPercent > 0 {
			top50pTasks := topK(tasks, fiftyPercent)
			medianTask := top50pTasks[0]
			report.Median = float32(medianTask.Duration) / float32(time.Millisecond)
			tenPercent := fiftyPercent / 5
			if tenPercent > 0 {
				top10pTasks := topK(tasks, tenPercent)
				task90th := top10pTasks[0]
				report.Top90th = float32(task90th.Duration) / float32(time.Millisecond)
				onePercent := tenPercent / 10
				if onePercent > 0 {
					top1pTasks := topK(top10pTasks, onePercent)
					task99pth := top1pTasks[0]
					report.Top99th = float32(task99pth.Duration) / float32(time.Millisecond)
					pointOnePercent := onePercent / 10
					if pointOnePercent > 0 {
						topPointOneTasks := topK(top1pTasks, pointOnePercent)
						task99Point9th := topPointOneTasks[0]
						report.Top99p9th = float32(task99Point9th.Duration) / float32(time.Millisecond)
					} else {
						report.Top99p9th = getTopDuration(top1pTasks)
					}
				} else {
					mostDuration := getTopDuration(top10pTasks)
					report.Top99th = mostDuration
					report.Top99p9th = mostDuration
				}
			} else {
				mostDuration := getTopDuration(tasks)
				report.Top90th = mostDuration
				report.Top99th = mostDuration
				report.Top99p9th = mostDuration
			}
		} else {
			mostDuration := getTopDuration(tasks)
			report.Median = mostDuration
			report.Top90th = mostDuration
			report.Top99th = mostDuration
			report.Top99p9th = mostDuration
		}
	}

	log(report)
}

func getTopDuration(tasks []Task) float32 {
	top := topK(tasks, 1)
	if len(top) < 1 {
		return 0
	}

	return float32(top[0].Duration) / float32(time.Millisecond)
}

func (c *metricsContainer) RemoveAll() any {
	tasks := c.tasks
	duration := c.duration
	drops := c.drops
	c.tasks = nil
	c.duration = 0
	c.drops = 0

	return tasksDurationPair{
		tasks:    tasks,
		duration: duration,
		drops:    drops,
	}
}

func log(report *StatReport) {
	writeReport(report)
	if logEnabled.True() {
		logx.Statf("(%s) 指标 [1m] - 请求: %.1f/s, 丢弃: %d, 平均耗时: %.1fms, 中位数: %.1fms, 90th: %.1fms, 99th: %.1fms, 99.9th: %.1fms",
			report.Name, report.ReqsPerSecond, report.Drops, report.Average, report.Median,
			report.Top90th, report.Top99th, report.Top99p9th)
	}
}

func writeReport(report *StatReport) {
	writeLock.Lock()
	defer writeLock.Unlock()

	if reportWriter != nil {
		if err := reportWriter.Write(report); err != nil {
			logx.Error(err)
		}
	}
}
