package cache

import (
	"fmt"
	"git.zc0901.com/go/god/lib/collection"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/proc"
	"git.zc0901.com/go/god/lib/stat"
	"git.zc0901.com/go/god/lib/stringx"
	"git.zc0901.com/go/god/lib/threading"
	"time"
)

const (
	cleanWorkers     = 5
	timingWheelSlots = 300
	taskKeyLen       = 8
)

var (
	timingWheel *collection.TimingWheel
	taskRunner  = threading.NewTaskRunner(cleanWorkers)
)

// 延迟清除缓存的任务
type delayTask struct {
	delay time.Duration
	task  func() error
	keys  []string
}

func init() {
	var err error
	timingWheel, err = collection.NewTimingWheel(time.Second, timingWheelSlots, clean)
	logx.Must(err)

	// 关闭程序时，先清楚缓存（）
	proc.AddShutdownListener(func() {
		timingWheel.Drain(clean)
	})
}

func clean(key, value interface{}) {
	taskRunner.Schedule(func() {
		dt := value.(delayTask)
		err := dt.task()
		if err != nil {
			return
		}

		// 如果失败则计算下一个执行时间
		next, ok := nextDelay(dt.delay)
		if ok {
			dt.delay = next
			timingWheel.SetTimer(key, dt, next)
		} else {
			msg := fmt.Sprintf("已重试但依然未能清除缓存: %q, error: %v", formatKeys(dt.keys), err)
			logx.Error(msg)
			stat.Report(msg)
		}
	})
}

func nextDelay(delay time.Duration) (time.Duration, bool) {
	switch delay {
	case time.Second:
		return time.Second * 5, true
	case time.Second * 5:
		return time.Minute, true
	case time.Minute:
		return time.Minute * 5, true
	case time.Minute * 5:
		return time.Hour, true
	default:
		return 0, false
	}
}

// AddCleanTask 增加默认一秒后执行的任务 task，清理指定的一组 keys
func AddCleanTask(task func() error, keys ...string) {
	timingWheel.SetTimer(stringx.Randn(taskKeyLen), delayTask{
		delay: time.Second,
		task:  task,
		keys:  keys,
	}, time.Second)
}
