package stat

import (
	"os"
	"sync"
	"time"

	"git.zc0901.com/go/god/lib/executors"

	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/syncx"
)

var (
	logInterval  = time.Minute
	writerLock   sync.Mutex
	reportWriter Writer = nil
	logEnabled          = syncx.ForAtomicBool(true)
)

type (
	// Writer 是一个定义 Write 方法的接口。
	Writer interface {
		Write(report *ReportItem) error
	}

	// ReportItem 是一个统计汇报项结构体。
	ReportItem struct {
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

	// Metrics 用于记录和汇报统计项。
	Metrics struct {
		dispatcher *executors.PeriodicalExecutor
		container  *metricsContainer
	}
)

// DisableLog 禁用统计日志。
func DisableLog() {
	logEnabled.Set(false)
}

func SetReportWriter(writer Writer) {
	writerLock.Lock()
	// 指定writer，通过writer将数据推送给 Prometheus Server
	reportWriter = writer
	writerLock.Unlock()
}

func NewMetrics(name string) *Metrics {
	manager := &metricsContainer{
		name: name,
		pid:  os.Getpid(),
	}

	return &Metrics{
		dispatcher: executors.NewPeriodicalExecutor(logInterval, manager),
		container:  manager,
	}
}

func (m *Metrics) SetName(name string) {
	m.dispatcher.Sync(func() {
		m.container.name = name
	})
}

func (m *Metrics) Add(task Task) {
	m.dispatcher.Add(task)
}

func (m *Metrics) AddDrop() {
	m.dispatcher.Add(Task{Drop: true})
}

type (
	metricsContainer struct {
		name     string
		pid      int
		tasks    []Task
		duration time.Duration
		drops    int
	}

	tasksDurationPair struct {
		tasks    []Task
		duration time.Duration
		drops    int
	}
)

func (c *metricsContainer) Add(v interface{}) bool {
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

// Execute 执行并写入远程普罗米修斯
func (c *metricsContainer) Execute(v interface{}) {
	pair := v.(tasksDurationPair)
	tasks := pair.tasks
	duration := pair.duration
	drops := pair.drops
	size := len(tasks)
	report := &ReportItem{
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
			top50Tasks := topK(tasks, fiftyPercent)
			medianTask := top50Tasks[0]
			report.Median = float32(medianTask.Duration) / float32(time.Millisecond)
			tenPercent := fiftyPercent / 5
			if tenPercent > 0 {
				top10pTasks := topK(tasks, tenPercent)
				task90th := top10pTasks[0]
				report.Top90th = float32(task90th.Duration) / float32(time.Millisecond)
				onePercent := tenPercent / 10
				if onePercent > 0 {
					top1pTasks := topK(top10pTasks, onePercent)
					task99th := top1pTasks[0]
					report.Top99th = float32(task99th.Duration) / float32(time.Millisecond)
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

func (c *metricsContainer) RemoveAll() interface{} {
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

func getTopDuration(tasks []Task) float32 {
	top := topK(tasks, 1)
	if len(top) < 1 {
		return 0
	}

	return float32(top[0].Duration) / float32(time.Millisecond)
}

// log 写入远程metric地址
func log(report *ReportItem) {
	writeReport(report)
	if logEnabled.True() {
		logx.Statf("(%s) - QPS: %.1f/s, 丢弃数: %d, 平均时长: %.1fms, 中位数: %.1fms, "+
			"90th: %.1fms, 99th: %.1fms, 99.9th: %.1fms",
			report.Name, report.ReqsPerSecond, report.Drops, report.Average, report.Median,
			report.Top90th, report.Top99th, report.Top99p9th)
	}
}

func writeReport(report *ReportItem) {
	writerLock.Lock()
	defer writerLock.Unlock()

	if reportWriter != nil {
		if err := reportWriter.Write(report); err != nil {
			logx.Error(err)
		}
	}
}
