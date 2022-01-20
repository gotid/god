package dq

import (
	"strconv"
	"time"

	"github.com/gotid/god/lib/hash"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/lib/store/redis"
)

const (
	expire     = 3600 // 秒
	guardValue = "1"
	tolerance  = time.Minute * 30
)

var maxCheckBytes = getMaxTimeLen()

type (
	Consume func(body []byte)

	Consumer interface {
		Consume(consume Consume)
	}

	consumerCluster struct {
		nodes []*consumerNode
		redis *redis.Redis
	}
)

// 新建消费者集群
func NewConsumer(c Conf) Consumer {
	var nodes []*consumerNode
	for _, node := range c.Beanstalks {
		nodes = append(nodes, newConsumerNode(node.Endpoint, node.Tube))
	}
	return &consumerCluster{
		nodes: nodes,
		redis: c.Redis.NewRedis(),
	}
}

// 消费任务
func (c *consumerCluster) Consume(consume Consume) {
	guardedConsume := func(body []byte) {
		key := hash.Md5Hex(body) // 以任务内容体为redis的key
		body, ok := c.unwrap(body)
		if !ok {
			logx.Errorf("丢弃队列任务：%q", string(body))
			return
		}

		ok, err := c.redis.SetNXEx(key, guardValue, expire) // 任务默认保留1小时
		if err != nil {
			logx.Error(err)
		} else if ok {
			consume(body)
		}
	}

	group := service.NewServiceGroup()
	for _, node := range c.nodes {
		group.Add(consumerService{
			node:    node,
			consume: guardedConsume,
		})
	}
	group.Start()
}

// 打开队列任务内容体
func (c *consumerCluster) unwrap(body []byte) ([]byte, bool) {
	pos := -1
	for i := 0; i < maxCheckBytes && i < len(body); i++ {
		if body[i] == timeSep {
			pos = i
			break
		}
	}
	if pos < 0 {
		return nil, false
	}

	val, err := strconv.ParseInt(string(body[:pos]), 10, 64)
	if err != nil {
		logx.Error(err)
		return nil, false
	}

	t := time.Unix(0, val)
	if t.Add(tolerance).Before(time.Now()) {
		return nil, false
	}

	return body[pos+1:], true
}

func getMaxTimeLen() int {
	return len(strconv.FormatInt(time.Now().UnixNano(), 10)) + 2
}
