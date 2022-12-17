package queue

type (
	// Producer 表示一个生产消息的生产者。
	Producer interface {
		// AddListener 添加一个生产者监听器
		AddListener(ProducerListener)

		// Produce 生产者进行生产
		Produce() (string, bool)
	}

	// ProducerListener 表示一个生产者监听器。
	ProducerListener interface {
		// OnProducerPause 暂停某个生产者的工作
		OnProducerPause()
		// OnProducerResume 恢复某个生产者的工作
		OnProducerResume()
	}

	// ProducerFactory 定义生成一个 Producer 的方法。
	ProducerFactory func() (Producer, error)
)
