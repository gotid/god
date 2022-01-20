package collection

import (
	"container/list"
	"fmt"
	"time"

	"git.zc0901.com/go/god/lib/lang"
	"git.zc0901.com/go/god/lib/threading"
	"git.zc0901.com/go/god/lib/timex"
)

const drainWorkers = 8

type (
	// TimingWheel 计时轮是用于调度任务的计时轮对象。
	TimingWheel struct {
		interval      time.Duration // 时间划分刻度，一个最小的时间格子，是定时器的转动单位
		ticker        timex.Ticker
		slots         []*list.List
		timers        *SafeMap
		numSlots      int     // 时间轮的插槽数量（有几个圆）
		tickedPos     int     // 定时器位置
		execute       Execute // 时间点执行函数
		setChannel    chan timingEntry
		moveChannel   chan baseEntry
		removeChannel chan interface{}
		drainChannel  chan func(key, value interface{})
		stopChannel   chan lang.PlaceholderType
	}

	Execute func(key, value interface{})

	baseEntry struct {
		delay time.Duration
		key   interface{}
	}

	timingEntry struct {
		baseEntry
		value   interface{}
		circle  int
		diff    int
		removed bool
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

func NewTimingWheel(interval time.Duration, numSlots int, execute Execute) (*TimingWheel, error) {
	if interval <= 0 || numSlots <= 0 || execute == nil {
		return nil, fmt.Errorf("执行间隔(%v)/插槽数量(%d) 必须大于零，执行函数不能为空(%p)", interval, numSlots, execute)
	}

	return newTimingWheelWithClock(interval, numSlots, execute, timex.NewTicker(interval))
}

// 真正做初始化
func newTimingWheelWithClock(interval time.Duration, numSlots int, execute Execute, ticker timex.Ticker) (*TimingWheel, error) {
	tw := &TimingWheel{
		interval:      interval,                                // 单个时间间隔
		ticker:        ticker,                                  // 定时器，做时间推动，以 interval 为单位推进
		slots:         make([]*list.List, numSlots),            // 时间槽，双向链表实现
		timers:        NewSafeMap(),                            // 存储 task{key, value} 的 map，提供 execute 执行函数所需参数
		numSlots:      numSlots,                                // 时间槽个数
		tickedPos:     numSlots - 1,                            // 位于上一次虚拟circle中
		execute:       execute,                                 // 时间点任务的真正执行函数
		setChannel:    make(chan timingEntry),                  // 设置任务的通道
		moveChannel:   make(chan baseEntry),                    // 移动任务的通道
		removeChannel: make(chan interface{}),                  // 移除任务的通道
		drainChannel:  make(chan func(key, value interface{})), // 执行任务通道
		stopChannel:   make(chan lang.PlaceholderType),         // 停止任务通道
	}

	tw.initSlots()
	go tw.run()

	return tw, nil
}

// Drain 排水：向排水通道发送排水函数
func (tw *TimingWheel) Drain(fn func(key, value interface{})) {
	tw.drainChannel <- fn
}

func (tw *TimingWheel) MoveTimer(key interface{}, delay time.Duration) {
	if delay <= 0 || key == nil {
		return
	}

	tw.moveChannel <- baseEntry{
		delay: delay,
		key:   key,
	}
}

func (tw *TimingWheel) RemoveTimer(key interface{}) {
	if key == nil {
		return
	}

	tw.removeChannel <- key
}

func (tw *TimingWheel) SetTimer(key, value interface{}, delay time.Duration) {
	if delay <= 0 || key == nil {
		return
	}

	tw.setChannel <- timingEntry{
		baseEntry: baseEntry{
			delay: delay,
			key:   key,
		},
		value: value,
	}
}

func (tw *TimingWheel) Stop() {
	close(tw.stopChannel)
}

func (tw *TimingWheel) initSlots() {
	for i := 0; i < tw.numSlots; i++ {
		tw.slots[i] = list.New()
	}
}

// 运行定时器，推动时间轮运转
func (tw *TimingWheel) run() {
	for {
		select {
		// 定时器推动时间
		case <-tw.ticker.Chan():
			tw.onTick()
		//	收到新任务
		case task := <-tw.setChannel:
			tw.setTask(&task)
		//	移除任务
		case key := <-tw.removeChannel:
			tw.removeTask(key)
		// 移动任务
		case task := <-tw.moveChannel:
			tw.moveTask(task)
		//	清洗任务
		case fn := <-tw.drainChannel:
			tw.drainAll(fn)
		// 停止定时器
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimingWheel) onTick() {
	tw.tickedPos = (tw.tickedPos + 1) % tw.numSlots
	l := tw.slots[tw.tickedPos]
	tw.scanAndRunTasks(l)
}

func (tw *TimingWheel) scanAndRunTasks(l *list.List) {
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
			pos := (tw.tickedPos + task.diff) % tw.numSlots
			tw.slots[pos].PushBack(task)
			tw.setTimerPosition(pos, task)
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
		tw.timers.Del(task.key)
		e = next
	}

	tw.runTasks(tasks)
}

func (tw *TimingWheel) setTimerPosition(pos int, task *timingEntry) {
	if val, ok := tw.timers.Get(task.key); ok {
		timer := val.(*positionEntry)
		timer.item = task
		timer.pos = pos
	} else {
		tw.timers.Set(task.key, &positionEntry{
			pos:  pos,
			item: task,
		})
	}
}

func (tw *TimingWheel) runTasks(tasks []timingTask) {
	if len(tasks) == 0 {
		return
	}

	go func() {
		for i := range tasks {
			threading.RunSafe(func() {
				tw.execute(tasks[i].key, tasks[i].value)
			})
		}
	}()
}

func (tw *TimingWheel) setTask(task *timingEntry) {
	if task.delay < tw.interval {
		task.delay = tw.interval
	}

	if val, ok := tw.timers.Get(task.key); ok {
		entry := val.(*positionEntry)
		entry.item.value = task.value
		tw.moveTask(task.baseEntry)
	} else {
		pos, circle := tw.getPositionAndCircle(task.delay)
		task.circle = circle
		tw.slots[pos].PushBack(task)
		tw.setTimerPosition(pos, task)
	}
}

func (tw *TimingWheel) moveTask(task baseEntry) {
	// 通过任务 key 名，获取 positionEntry 位置实体信息（位置信息、任务信息）
	val, ok := tw.timers.Get(task.key)
	if !ok {
		return
	}

	timer := val.(*positionEntry)
	// 任务的延迟时间比时间格间隔还要小，说明应该立即执行
	if task.delay < tw.interval {
		threading.RunSafe(func() {
			tw.execute(timer.item.key, timer.item.value)
		})
		return
	}

	pos, circle := tw.getPositionAndCircle(task.delay)
	// 如果比时间格间隔大，则通过延迟时间算出时间轮中的新位置pos和circle
	if pos >= timer.pos {
		timer.item.circle = circle
		// 记录前后移动的偏移量，是为了后续重新入队
		timer.item.diff = pos - timer.pos
	} else if circle > 0 {
		// 移动到下一个内圆，将 circle 转换为 diff 的一部分
		circle--
		timer.item.circle = circle
		// 因为是一个数组，要加上 numSlots（相当于要走到下一层？）
		timer.item.diff = tw.numSlots + pos - timer.pos
	} else {
		// 如果 offset 提前了，此时 task 也还在第一层
		// 标记删除老的 task，并重新入队，等待被执行
		timer.item.removed = true
		newTask := &timingEntry{
			baseEntry: task,
			value:     timer.item.value,
		}
		tw.slots[pos].PushBack(newTask)
		tw.setTimerPosition(pos, newTask)
	}
}

func (tw *TimingWheel) getPositionAndCircle(delay time.Duration) (pos int, circle int) {
	steps := int(delay / tw.interval)
	pos = (tw.tickedPos + steps) % tw.numSlots
	circle = (steps - 1) / tw.numSlots

	return
}

func (tw *TimingWheel) removeTask(key interface{}) {
	val, ok := tw.timers.Get(key)
	if !ok {
		return
	}

	timer := val.(*positionEntry)
	timer.item.removed = true
	tw.timers.Del(key)
}

func (tw *TimingWheel) drainAll(fn func(key interface{}, value interface{})) {
	runner := threading.NewTaskRunner(drainWorkers)
	for _, slot := range tw.slots {
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
