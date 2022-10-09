package cache

import (
	"fmt"
	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/stringx"
	"github.com/gotid/god/lib/threading"
	"time"
)

const (
	timingWheelSlots = 300
	cleanWorkers     = 5
	taskKeyLen       = 8
)

var (
	timingWheel *collection.TimingWheel
	taskRunner  = threading.NewTaskRunner(cleanWorkers)
)

type delayTask struct {
	delay time.Duration
	task  func() error
	keys  []string
}

func init() {
	var err error
	timingWheel, err = collection.NewTimingWheel(time.Second, timingWheelSlots, clean)
	logx.Must(err)

	proc.AddShutdownListener(func() {
		timingWheel.Drain(clean)
	})
}

// AddCleanTask 添加一个清洗任务。
func AddCleanTask(task func() error, keys ...string) {
	timingWheel.SetTimer(stringx.Randn(taskKeyLen), delayTask{
		delay: time.Second,
		task:  task,
		keys:  keys,
	}, time.Second)
}

func clean(key, val interface{}) {
	taskRunner.Schedule(func() {
		dt := val.(delayTask)
		err := dt.task()
		if err != nil {
			return
		}

		next, ok := nextDelay(dt.delay)
		if ok {
			dt.delay = next
			timingWheel.SetTimer(key, dt, next)
		} else {
			msg := fmt.Sprintf("已获取但清除缓存失败，键：%q，错误：%v",
				formatKeys(dt.keys), err)
			logx.Error(msg)
			stat.Report(msg)
		}
	})
}

func nextDelay(delay time.Duration) (time.Duration, bool) {
	switch delay {
	case time.Second:
		return 5 * time.Second, true
	case 5 * time.Second:
		return time.Minute, true
	case time.Minute:
		return 5 * time.Minute, true
	case 5 * time.Minute:
		return time.Hour, true
	default:
		return 0, false
	}
}
