package queue

// MessageQueue 表示一个消息队列。
type MessageQueue interface {
	Start()
	Stop()
}
