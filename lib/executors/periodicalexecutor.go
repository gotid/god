package executors

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/threading"
	"github.com/gotid/god/lib/timex"
)

const idleRound = 10

type (
	// TaskContainer 任务容器：负责任务的新增、执行、移除。
	TaskContainer interface {
		Add(task interface{}) bool // 添加任务
		Execute(tasks interface{}) // 执行任务
		RemoveAll() interface{}    // 删除并返回当前所有任务
	}

	// PeriodicalExecutor 定时调度器
	PeriodicalExecutor struct {
		interval    time.Duration                             // 任务间隔
		container   TaskContainer                             // 任务容器
		commander   chan interface{}                          // 任务通道
		confirmChan chan lang.PlaceholderType                 // 任务确认通道
		waitGroup   sync.WaitGroup                            // 同步等待组
		wgLocker    syncx.Locker                              // 同步等待组的加锁器
		inflight    int32                                     // 飞行中协程数
		guarded     bool                                      // 是否守卫
		newTicker   func(duration time.Duration) timex.Ticker // 任务断续器
		lock        sync.Mutex
	}
)

// NewPeriodicalExecutor 定时调度器（间隔时间，任务管理器）
func NewPeriodicalExecutor(interval time.Duration, container TaskContainer) *PeriodicalExecutor {
	executor := &PeriodicalExecutor{
		interval:    interval,
		container:   container,
		commander:   make(chan interface{}, 1),
		confirmChan: make(chan lang.PlaceholderType),
		newTicker: func(duration time.Duration) timex.Ticker {
			return timex.NewTicker(duration)
		},
	}

	// 程序关闭前，尽量执行剩余任务
	proc.AddShutdownListener(func() {
		executor.Flush()
	})

	return executor
}

// Add 添加新任务给任务通道并确认可以执行
func (pd *PeriodicalExecutor) Add(task interface{}) {
	if tasks, ok := pd.setAndGet(task); ok {
		pd.commander <- tasks // 将当前所有任务发给任务通道
		<-pd.confirmChan      // 确认通道进行确认
	}
}

// Flush 清洗任务
func (pd *PeriodicalExecutor) Flush() bool {
	pd.enter()
	return pd.execute(func() interface{} {
		pd.lock.Lock()
		defer pd.lock.Unlock()
		return pd.container.RemoveAll()
	}())
}

// Sync 同步执行一个自定义函数
func (pd *PeriodicalExecutor) Sync(fn func()) {
	pd.lock.Lock()
	defer pd.lock.Unlock()
	fn()
}

// Wait 加锁保护等待操作
func (pd *PeriodicalExecutor) Wait() {
	pd.Flush()
	pd.wgLocker.Guard(func() {
		pd.waitGroup.Wait()
	})
}

// setAndGet 新增并返回任务，如有可能则后台直接执行任务
// 返回：加入后的所有待处理任务，是否已递交任务管理者
func (pd *PeriodicalExecutor) setAndGet(task interface{}) (interface{}, bool) {
	pd.lock.Lock()
	defer func() {
		if !pd.guarded {
			pd.guarded = true
			// 快速解锁
			defer pd.backgroundFlush()
		}
		pd.lock.Unlock()
	}()

	if pd.container.Add(task) {
		atomic.AddInt32(&pd.inflight, 1)
		return pd.container.RemoveAll(), true
	}

	return nil, false
}

// 后台任务清洗 - 起一个协程来处理任务
func (pd *PeriodicalExecutor) backgroundFlush() {
	threading.GoSafe(func() {
		// 退出协程之前进行清理，避免丢失任务
		defer pd.Flush()

		ticker := pd.newTicker(pd.interval)
		defer ticker.Stop()

		// 任务通道调度定时执行器
		var executed bool
		lastTime := timex.Now()
		for {
			select {
			case tasks := <-pd.commander: // 主动触发上报
				executed = true
				atomic.AddInt32(&pd.inflight, -1)
				pd.enter()
				pd.confirmChan <- lang.Placeholder
				pd.execute(tasks)
				lastTime = timex.Now()
			case <-ticker.Chan(): // 定时上报
				if executed {
					executed = false
				} else if pd.Flush() {
					lastTime = timex.Now()
				} else if pd.shallQuit(lastTime) {
					return
				}
			}
		}
	})
}

// enter 执行者进入，等待组加锁
func (pd *PeriodicalExecutor) enter() {
	pd.wgLocker.Guard(func() {
		pd.waitGroup.Add(1)
	})
}

// execute 调度任务管理者，执行任务
func (pd *PeriodicalExecutor) execute(tasks interface{}) bool {
	defer pd.done()

	ok := pd.has(tasks)
	if ok {
		pd.container.Execute(tasks)
	}

	return ok
}

// done 完成分组任务
func (pd *PeriodicalExecutor) done() {
	pd.waitGroup.Done()
}

// has 判断任务有无
func (pd *PeriodicalExecutor) has(tasks interface{}) bool {
	if tasks == nil {
		return false
	}

	val := reflect.ValueOf(tasks)
	switch val.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return val.Len() > 0
	default:
		// 其他类型默认为有值，让调用者自行处理
		return true
	}
}

func (pd *PeriodicalExecutor) shallQuit(lastTime time.Duration) (stop bool) {
	if timex.Since(lastTime) <= pd.interval*idleRound {
		return
	}

	pd.lock.Lock()
	if atomic.LoadInt32(&pd.inflight) == 0 {
		pd.guarded = false
		stop = true
	}
	pd.lock.Unlock()

	return
}
