package executors

import (
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/threading"
	"github.com/gotid/god/lib/timex"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

const idleRound = 10

type (
	// TaskContainer 接口定义了一个可用作底层容器执行定时任务的类型。
	TaskContainer interface {
		// AddTask 添加任务值容器。
		// 如果添加后需要刷新则返回真。
		AddTask(task interface{}) bool
		// Execute 在刷新时处理容器收集的任务。
		Execute(tasks interface{})
		// RemoveAll 移除容器中的所有任务并返回它们。
		RemoveAll() interface{}
	}

	// PeriodicalExecutor 是一个定期执行器。
	PeriodicalExecutor struct {
		commander chan interface{}
		interval  time.Duration
		container TaskContainer
		waitGroup sync.WaitGroup
		// 避免 WaitGroup 调用 Add/Done/Wait 时出现竞争条件
		wgBarrier   syncx.Barrier
		confirmChan chan lang.PlaceholderType
		inflight    int32
		guarded     bool
		newTicker   func(duration time.Duration) timex.Ticker
		lock        sync.Mutex
	}
)

// NewPeriodicalExecutor 使用给定的 interval 和 container，返回一个 PeriodicalExecutor。
func NewPeriodicalExecutor(interval time.Duration, container TaskContainer) *PeriodicalExecutor {
	executor := &PeriodicalExecutor{
		// 缓冲为1以提高调用速度
		commander:   make(chan interface{}, 1),
		interval:    interval,
		container:   container,
		confirmChan: make(chan lang.PlaceholderType),
		newTicker: func(d time.Duration) timex.Ticker {
			return timex.NewTicker(d)
		},
	}
	proc.AddShutdownListener(func() {
		executor.Flush()
	})

	return executor
}

// Add 添加任务到 pe。
func (pe *PeriodicalExecutor) Add(task interface{}) {
	if values, ok := pe.addAndCheck(task); ok {
		pe.commander <- values
		<-pe.confirmChan
	}
}

// Flush 强制 pe 执行任务。
func (pe *PeriodicalExecutor) Flush() bool {
	pe.enterExecution()
	return pe.executeTasks(func() interface{} {
		pe.lock.Lock()
		defer pe.lock.Unlock()
		return pe.container.RemoveAll()
	}())
}

// Sync 允许调用者使用 pe 执行线程安全的 fn 调用，尤其是底层容器。
func (pe *PeriodicalExecutor) Sync(fn func()) {
	pe.lock.Lock()
	defer pe.lock.Unlock()
	fn()
}

// Wait 等待执行完成。
func (pe *PeriodicalExecutor) Wait() {
	pe.Flush()
	pe.wgBarrier.Guard(func() {
		pe.waitGroup.Wait()
	})
}

func (pe *PeriodicalExecutor) enterExecution() {
	pe.wgBarrier.Guard(func() {
		pe.waitGroup.Add(1)
	})
}

func (pe *PeriodicalExecutor) executeTasks(tasks interface{}) bool {
	defer pe.doneExecution()

	ok := pe.hasTasks(tasks)
	if ok {
		pe.container.Execute(tasks)
	}

	return ok
}

func (pe *PeriodicalExecutor) doneExecution() {
	pe.waitGroup.Done()
}

func (pe *PeriodicalExecutor) hasTasks(tasks interface{}) bool {
	if tasks == nil {
		return false
	}

	val := reflect.ValueOf(tasks)
	switch val.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return val.Len() > 0
	default:
		// 未知类型，让调用者执行
		return true
	}
}

func (pe *PeriodicalExecutor) addAndCheck(task interface{}) (interface{}, bool) {
	pe.lock.Lock()
	defer func() {
		if !pe.guarded {
			pe.guarded = true
			defer pe.backgroundFlush()
		}
		pe.lock.Unlock()
	}()

	if pe.container.AddTask(task) {
		atomic.AddInt32(&pe.inflight, 1)
		return pe.container.RemoveAll(), true
	}

	return nil, false
}

func (pe *PeriodicalExecutor) backgroundFlush() {
	threading.GoSafe(func() {
		// 退出协程前进行刷新以避免丢失任务
		defer pe.Flush()

		ticker := pe.newTicker(pe.interval)
		defer ticker.Stop()

		var commanded bool
		last := timex.Now()
		for {
			select {
			case tasks := <-pe.commander:
				commanded = true
				atomic.AddInt32(&pe.inflight, -1)
				pe.enterExecution()
				pe.confirmChan <- lang.Placeholder
				pe.executeTasks(tasks)
				last = timex.Now()
			case <-ticker.Chan():
				if commanded {
					commanded = false
				} else if pe.Flush() {
					last = timex.Now()
				} else if pe.shallQuit(last) {
					return
				}
			}
		}
	})
}

func (pe *PeriodicalExecutor) shallQuit(last time.Duration) (stop bool) {
	if timex.Since(last) <= pe.interval*idleRound {
		return
	}

	// 检查 pe.inflight 和设置 pe.guarded 应当一起加锁
	pe.lock.Lock()
	if atomic.LoadInt32(&pe.inflight) == 0 {
		pe.guarded = false
		stop = true
	}
	pe.lock.Unlock()

	return
}
