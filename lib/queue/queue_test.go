package queue

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	consumers = 4
	rounds    = 100
)

func TestQueue(t *testing.T) {
	producer := newMockedProducer(rounds)
	consumer := newMockedConsumer()
	consumer.wait.Add(consumers)

	q := NewQueue(func() (Producer, error) {
		return producer, nil
	}, func() (Consumer, error) {
		return consumer, nil
	})
	q.AddListener(new(mockedListener))
	q.SetName("mockQueue")
	q.SetConsumerCount(consumers)
	q.SetProducerCount(1)
	q.pause()
	q.resume()
	go func() {
		producer.wait.Wait()
		q.Stop()
	}()
	q.Start()
	assert.Equal(t, int32(rounds), atomic.LoadInt32(&consumer.count))
}

type mockedListener struct {
}

func (l *mockedListener) OnPause() {
	println("监听器暂停...")
}

func (l *mockedListener) OnResume() {
	println("监听器恢复...")
}

type mockedProducer struct {
	total int32
	count int32
	wait  sync.WaitGroup
}

func (p *mockedProducer) AddListener(_ ProducerListener) {

}

func (p *mockedProducer) Produce() (string, bool) {
	if atomic.AddInt32(&p.count, 1) <= p.total {
		p.wait.Done()
		return fmt.Sprintf("item-round-%d", atomic.LoadInt32(&p.count)), true
	}

	time.Sleep(time.Second)
	return "", false
}

func newMockedProducer(total int32) *mockedProducer {
	p := new(mockedProducer)
	p.total = total
	p.wait.Add(int(total))
	return p
}

type mockedConsumer struct {
	count  int32
	events int32
	wait   sync.WaitGroup
}

func (c *mockedConsumer) Consume(message string) error {
	atomic.AddInt32(&c.count, 1)
	fmt.Println("消费者消费消息：", message)
	return nil
}

func (c *mockedConsumer) OnEvent(event any) {
	fmt.Println("消费者处理事件：", event)
	if atomic.AddInt32(&c.events, 1) <= consumers {
		c.wait.Done()
	}
}

func newMockedConsumer() *mockedConsumer {
	return new(mockedConsumer)
}
