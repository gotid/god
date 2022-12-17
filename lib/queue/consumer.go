package queue

type (
	// Consumer 表示一个消费字符串消息的消费者。
	Consumer interface {
		// Consume 消费者进行消费
		Consume(string) error
		// OnEvent 消费者处理事件
		OnEvent(event any)
	}

	// ConsumerFactory 定义生成消费者的方法。
	ConsumerFactory func() (Consumer, error)
)
