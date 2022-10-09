package collection

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/threading"
	"github.com/gotid/god/lib/timex"
	"time"
)

const drainWorkers = 8

var (
	ErrClosed   = errors.New("时间轮已关闭")
	ErrArgument = errors.New("错误的任务参数")
)

type (
	// Execute 定义执行任务的方法。
	Execute func(key, val interface{})

	// TimingWheel 是一个调度任务的时间轮对象。
	TimingWheel struct {
		interval      time.Duration
		ticker        timex.Ticker
		slots         []*list.List
		timers        *SafeMap
		tickedPos     int
		numSlots      int
		execute       Execute
		setChannel    chan timingEntry
		moveChannel   chan baseEntry
		removeChannel chan interface{}
		drainChannel  chan func(key, value interface{})
		stopChannel   chan lang.PlaceholderType
	}

	timingEntry struct {
		baseEntry
		value   interface{}
		circle  int
		diff    int
		removed bool
	}

	baseEntry struct {
		delay time.Duration
		key   interface{}
	}

	positionEntry struct {
		pos  int
		item *timingEntry
	}

	timingTask struct {
		key   interface{}
		value interface{}
	}
)

// NewTimingWheel 返回一个时间轮 TimingWheel。
func NewTimingWheel(interval time.Duration, numSlots int, execute Execute) (*TimingWheel, error) {
	if interval <= 0 || numSlots <= 0 || execute == nil {
		return nil, fmt.Errorf("间隔：%v，槽位：%d，执行函数：%p",
			interval, numSlots, execute)
	}

	return newTimingWheelWithClock(interval, numSlots, execute, timex.NewTicker(interval))
}

// Drain 排干所有项目并执行它们。
func (w *TimingWheel) Drain(fn func(key, value interface{})) error {
	select {
	case w.drainChannel <- fn:
		return nil
	case <-w.stopChannel:
		return ErrClosed
	}
}

// MoveTimer 将给定键的任务移动到给定延迟。
func (w *TimingWheel) MoveTimer(key interface{}, delay time.Duration) error {
	if delay <= 0 || key == nil {
		return ErrArgument
	}

	select {
	case w.moveChannel <- baseEntry{
		delay: delay,
		key:   key,
	}:
		return nil
	case <-w.stopChannel:
		return ErrClosed
	}
}

// RemoveTimer 移除给定键的任务。
func (w *TimingWheel) RemoveTimer(key interface{}) error {
	if key == nil {
		return ErrArgument
	}

	select {
	case w.removeChannel <- key:
		return nil
	case <-w.stopChannel:
		return ErrClosed
	}
}

// SetTimer 设置键值及其延迟执行时间。
func (w *TimingWheel) SetTimer(key, value interface{}, delay time.Duration) error {
	if delay <= 0 || key == nil {
		return ErrArgument
	}

	select {
	case w.setChannel <- timingEntry{
		baseEntry: baseEntry{
			delay: delay,
			key:   key,
		},
		value: value,
	}:
		return nil
	case <-w.stopChannel:
		return ErrClosed
	}
}

// Stop 停止时间轮。
func (w *TimingWheel) Stop() {
	close(w.stopChannel)
}

func newTimingWheelWithClock(interval time.Duration, numSlots int, execute Execute, ticker timex.Ticker) (*TimingWheel, error) {
	tw := &TimingWheel{
		interval:      interval,
		ticker:        ticker,
		slots:         make([]*list.List, numSlots),
		timers:        NewSafeMap(),
		tickedPos:     numSlots - 1,
		numSlots:      numSlots,
		execute:       execute,
		setChannel:    make(chan timingEntry),
		moveChannel:   make(chan baseEntry),
		removeChannel: make(chan interface{}),
		drainChannel:  make(chan func(key, value interface{})),
		stopChannel:   make(chan lang.PlaceholderType),
	}

	tw.initSlots()
	go tw.run()

	return tw, nil
}

func (w *TimingWheel) initSlots() {
	for i := 0; i < w.numSlots; i++ {
		w.slots[i] = list.New()
	}
}

func (w *TimingWheel) run() {
	for {
		select {
		case <-w.ticker.Chan():
			w.onTick()
		case task := <-w.setChannel:
			w.setTask(&task)
		case key := <-w.removeChannel:
			w.removeTask(key)
		case task := <-w.moveChannel:
			w.moveTask(task)
		case fn := <-w.drainChannel:
			w.drainAll(fn)
		case <-w.stopChannel:
			w.ticker.Stop()
			return
		}
	}
}

func (w *TimingWheel) onTick() {
	w.tickedPos = (w.tickedPos + 1) % w.numSlots
	l := w.slots[w.tickedPos]
	w.scanAndRunTasks(l)
}

func (w *TimingWheel) scanAndRunTasks(l *list.List) {
	var tasks []timingTask

	for e := l.Front(); e != nil; {
		task := e.Value.(*timingEntry)
		if task.removed {
			next := e.Next()
			l.Remove(e)
			e = next
			continue
		} else if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		} else if task.diff > 0 {
			next := e.Next()
			l.Remove(e)
			pos := (w.tickedPos + task.diff) % w.numSlots
			w.slots[pos].PushBack(task)
			w.setTimerPosition(pos, task)
			task.diff = 0
			e = next
			continue
		}

		tasks = append(tasks, timingTask{
			key:   task.key,
			value: task.value,
		})
		next := e.Next()
		l.Remove(e)
		w.timers.Del(task.key)
		e = next
	}

	w.runTasks(tasks)
}

func (w *TimingWheel) setTimerPosition(pos int, task *timingEntry) {
	if val, ok := w.timers.Get(task.key); ok {
		timer := val.(*positionEntry)
		timer.item = task
		timer.pos = pos
	} else {
		w.timers.Put(task.key, &positionEntry{
			pos:  pos,
			item: task,
		})
	}
}

func (w *TimingWheel) runTasks(tasks []timingTask) {
	if len(tasks) == 0 {
		return
	}

	go func() {
		for i := range tasks {
			threading.RunSafe(func() {
				w.execute(tasks[i].key, tasks[i].value)
			})
		}
	}()
}

func (w *TimingWheel) setTask(task *timingEntry) {
	if task.delay < w.interval {
		task.delay = w.interval
	}

	if val, ok := w.timers.Get(task.key); ok {
		entry := val.(*positionEntry)
		entry.item.value = task.value
		w.moveTask(task.baseEntry)
	} else {
		pos, circle := w.getPositionAndCircle(task.delay)
		task.circle = circle
		w.slots[pos].PushBack(task)
		w.setTimerPosition(pos, task)
	}
}

func (w *TimingWheel) moveTask(task baseEntry) {
	val, ok := w.timers.Get(task.key)
	if !ok {
		return
	}

	timer := val.(*positionEntry)
	if task.delay < w.interval {
		threading.GoSafe(func() {
			w.execute(timer.item.key, timer.item.value)
		})
		return
	}

	pos, circle := w.getPositionAndCircle(task.delay)
	if pos > timer.pos {
		timer.item.circle = circle
		timer.item.diff = pos - timer.pos
	} else if circle > 0 {
		circle--
		timer.item.circle = circle
		timer.item.diff = w.numSlots + pos - timer.pos
	} else {
		timer.item.removed = true
		newItem := &timingEntry{
			baseEntry: task,
			value:     timer.item.value,
		}
		w.slots[pos].PushBack(newItem)
		w.setTimerPosition(pos, newItem)
	}
}

func (w *TimingWheel) getPositionAndCircle(d time.Duration) (pos, circle int) {
	steps := int(d / w.interval)
	pos = (w.tickedPos + steps) % w.numSlots
	circle = (steps - 1) / w.numSlots

	return
}

func (w *TimingWheel) removeTask(key interface{}) {
	val, ok := w.timers.Get(key)
	if !ok {
		return
	}

	timer := val.(*positionEntry)
	timer.item.removed = true
	w.timers.Del(key)
}

func (w *TimingWheel) drainAll(fn func(key, value interface{})) {
	runner := threading.NewTaskRunner(drainWorkers)
	for _, slot := range w.slots {
		for e := slot.Front(); e != nil; {
			task := e.Value.(*timingEntry)
			next := e.Next()
			slot.Remove(e)
			e = next
			if !task.removed {
				runner.Schedule(func() {
					fn(task.key, task.value)
				})
			}
		}
	}
}
