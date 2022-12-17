package queue

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/rescue"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/threading"
	"github.com/gotid/god/lib/timex"
)

const queueName = "queue"

type (
	// Queue 是一个消息队列。
	Queue struct {
		name                 string
		metrics              *stat.Metrics
		producerFactory      ProducerFactory         // 生产者工厂
		producerRoutineGroup *threading.RoutineGroup // 生产者协程组
		consumerFactory      ConsumerFactory         // 消费者工厂
		consumerRoutineGroup *threading.RoutineGroup // 消费者协程组
		producerCount        int                     // 生产者个数
		consumerCount        int                     // 消费者个数
		active               int32                   // 活跃的生产者个数
		channel              chan string             // 消息通道
		quit                 chan struct{}           // 关闭通道
		listeners            []Listener              // 队列监听器
		eventLock            sync.Mutex              // 事件读写互斥锁
		eventChannels        []chan any              // 事件通道
	}

	// Listener 表示一个监听队列事件的监听器。
	Listener interface {
		OnPause()
		OnResume()
	}

	// Poller 拉手，包装 Poll 方法的接口。
	Poller interface {
		Name() string // 拉手名称
		Poll() string // 拉取消息
	}

	// Pusher 推手，包装 Push 方法的接口。
	Pusher interface {
		Name() string      // 推手名称
		Push(string) error // 推送消息
	}
)

// NewQueue 返回一个 Queue。
func NewQueue(producerFactory ProducerFactory, consumerFactory ConsumerFactory) *Queue {
	q := &Queue{
		metrics:              stat.NewMetrics(queueName),
		producerFactory:      producerFactory,
		producerRoutineGroup: threading.NewRoutineGroup(),
		consumerFactory:      consumerFactory,
		consumerRoutineGroup: threading.NewRoutineGroup(),
		producerCount:        runtime.NumCPU(),
		consumerCount:        runtime.NumCPU() << 1,
		channel:              make(chan string),
		quit:                 make(chan struct{}),
	}
	q.SetName(queueName)

	return q
}

// AddListener 添加一个队列监听器。
func (q *Queue) AddListener(listener Listener) {
	q.listeners = append(q.listeners, listener)
}

// Broadcast 向所有事件通道广播消息。
func (q *Queue) Broadcast(message any) {
	go func() {
		q.eventLock.Lock()
		defer q.eventLock.Unlock()

		for _, channel := range q.eventChannels {
			channel <- message
		}
	}()
}

// SetName 设置队列名称。
func (q *Queue) SetName(name string) {
	q.name = name
	q.metrics.SetName(name)
}

// SetConsumerCount 设置消费者个数。
func (q *Queue) SetConsumerCount(count int) {
	q.consumerCount = count
}

// SetProducerCount 设置生产者个数。
func (q *Queue) SetProducerCount(count int) {
	q.producerCount = count
}

// Start 启动队列。
func (q *Queue) Start() {
	q.startProducers(q.producerCount)
	q.startConsumers(q.consumerCount)

	q.producerRoutineGroup.Wait()
	close(q.channel)
	q.consumerRoutineGroup.Wait()
}

// Stop 停止队列。
func (q *Queue) Stop() {
	close(q.quit)
}

func (q *Queue) pause() {
	for _, listener := range q.listeners {
		listener.OnPause()
	}
}

func (q *Queue) resume() {
	for _, listener := range q.listeners {
		listener.OnResume()
	}
}

func (q *Queue) startProducers(count int) {
	for i := 0; i < count; i++ {
		q.producerRoutineGroup.Run(func() {
			q.produce()
		})
	}
}

func (q *Queue) produce() {
	// 创建一个可用的生产者
	var producer Producer
	for {
		var err error
		if producer, err = q.producerFactory(); err != nil {
			logx.Errorf("创建生产者失败：%v", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	// 向队列增加活跃的生产者
	atomic.AddInt32(&q.active, 1)

	// 向生产者增加要监控的队列
	producer.AddListener(routineListener{
		queue: q,
	})

	// 生产者读取通道（要么退出，要么生产）
	for {
		select {
		case <-q.quit:
			logx.Info("生产者接收到：队列退出")
			return
		default:
			if v, ok := q.produceOne(producer); ok {
				q.channel <- v
			}
		}
	}
}

func (q *Queue) produceOne(producer Producer) (string, bool) {
	// 避免 panic 退出生产者，只做记录并继续工作
	defer rescue.Recover()

	return producer.Produce()
}

func (q *Queue) startConsumers(count int) {
	// 一个消费者，对应一个事件通道
	for i := 0; i < count; i++ {
		eventChan := make(chan any)
		q.eventLock.Lock()
		q.eventChannels = append(q.eventChannels, eventChan)
		q.eventLock.Unlock()
		q.consumerRoutineGroup.Run(func() {
			q.consume(eventChan)
		})
	}
}

func (q *Queue) consume(eventChan chan any) {
	// 创建一个可用的消费者
	var consumer Consumer
	for {
		var err error
		if consumer, err = q.consumerFactory(); err != nil {
			logx.Errorf("创建消费者失败：%v", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	// 消费者读取通道，进行消费或处理事件
	for {
		select {
		case message, ok := <-q.channel:
			if ok {
				q.consumeOne(consumer, message)
			} else {
				logx.Info("任务队列已关闭，退出该消费者...")
				return
			}
		case event := <-eventChan:
			consumer.OnEvent(event)
		}
	}
}

func (q *Queue) consumeOne(consumer Consumer, message string) {
	threading.RunSafe(func() {
		start := timex.Now()
		defer func() {
			duration := timex.Since(start)
			q.metrics.Add(stat.Task{
				Duration: duration,
			})
			logx.WithDuration(duration).Info(message)
		}()

		if err := consumer.Consume(message); err != nil {
			logx.Errorf("错误发生在消费 %v：%v", message, err)
		}
	})
}

// 生产者的协程监听器
type routineListener struct {
	queue *Queue
}

func (l routineListener) OnProducerPause() {
	// 暂停某个生产者工作时，如果没有活跃工人了，则暂停队列
	if atomic.AddInt32(&l.queue.active, -1) <= 0 {
		l.queue.pause()
	}
}

func (l routineListener) OnProducerResume() {
	// 恢复某个生产者工作时，如果正好是第一个活跃工人，则恢复队列
	if atomic.AddInt32(&l.queue.active, 1) == 1 {
		l.queue.resume()
	}
}
