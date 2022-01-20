package dq

import (
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
)

type (
	// 消费者节点
	consumerNode struct {
		conn *connection
		tube string
		on   *syncx.AtomicBool
	}

	// 消费者服务
	consumerService struct {
		node    *consumerNode
		consume Consume
	}
)

// 新建消费者节点
func newConsumerNode(endpoint, tube string) *consumerNode {
	return &consumerNode{
		conn: newConnection(endpoint, tube),
		tube: tube,
		on:   syncx.ForAtomicBool(true),
	}
}

// 消费事件
func (n *consumerNode) consumeEvents(consume Consume) {
	for n.on.True() {
		conn, err := n.conn.get()
		if err != nil {
			logx.Error(err)
			time.Sleep(time.Second)
			continue
		}

		// 因为获取队列任务至多一秒，预订任务至多5秒，
		// 如果我们这里不检查 on/off 开关状态，那么连接可能不能关闭，因为
		// 平滑重启至多等待 5.5 秒。
		if !n.on.True() {
			break
		}

		// 使用指定管道，并预定其中的任务
		conn.Tube.Name = n.tube
		conn.TubeSet.Name[n.tube] = true
		id, body, err := conn.Reserve(reserveTimeout)
		if err == nil {
			conn.Delete(id)
			consume(body)
			continue
		}

		// 错误只允许是 beanstalk.NameError 或 beanstalk.ConnError
		switch e := err.(type) {
		case beanstalk.ConnError:
			switch e.Err {
			case beanstalk.ErrTimeout:
			// 超时，仅需继续循环，等待下次尝试
			case beanstalk.ErrBadChar, beanstalk.ErrBadFormat, beanstalk.ErrBuried, beanstalk.ErrDeadline,
				beanstalk.ErrDraining, beanstalk.ErrEmpty, beanstalk.ErrInternal, beanstalk.ErrJobTooBig,
				beanstalk.ErrNoCRLF, beanstalk.ErrNotFound, beanstalk.ErrNotIgnored, beanstalk.ErrTooLong:
				// 上述错误不会重置，需要记录错误
				logx.Error(err)
			default:
				// beanstalk.ErrOOM, beanstalk.ErrUnknown 和其他错误
				logx.Error(err)
				n.conn.reset()
				time.Sleep(time.Second)
			}
		default:
			logx.Error(err)
		}
	}

	if err := n.conn.Close(); err != nil {
		logx.Error(err)
	}
}

// 销毁节点
func (n *consumerNode) dispose() {
	n.on.Set(false)
}

// 开启消费者服务
func (c consumerService) Start() {
	c.node.consumeEvents(c.consume)
}

// 停止消费者服务
func (c consumerService) Stop() {
	c.node.dispose()
}
