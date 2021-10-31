package dq

import (
	"bytes"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"git.zc0901.com/go/god/lib/errorx"
	"git.zc0901.com/go/god/lib/fx"
	"git.zc0901.com/go/god/lib/lang"
	"git.zc0901.com/go/god/lib/logx"
)

const (
	replicaNodes    = 3 // 默认副本节点数
	minWrittenNodes = 2 // 最少可写节点数
)

type (
	// Producer 任务生产者
	Producer interface {
		At(body []byte, at time.Time) (string, error)           // 定时执行
		Delay(body []byte, delay time.Duration) (string, error) // 延迟执行
		Revoke(ids string) error                                // 撤回任务
		Close() error
	}

	// 生产者集群
	producerCluster struct {
		nodes []Producer
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewProducer 新建队列生产者集群
func NewProducer(beanstalks []Beanstalk) Producer {
	if len(beanstalks) < minWrittenNodes {
		log.Fatalf("节点数必须大于等于 %d", minWrittenNodes)
	}

	var nodes []Producer
	producers := make(map[string]lang.PlaceholderType)
	for _, node := range beanstalks {
		if _, ok := producers[node.Endpoint]; ok {
			log.Fatal("Beanstalk 节点地址不能重复")
		}

		producers[node.Endpoint] = lang.Placeholder
		nodes = append(nodes, NewProducerNode(node.Endpoint, node.Tube))
	}
	return &producerCluster{nodes: nodes}
}

// At 定时执行
func (p *producerCluster) At(body []byte, at time.Time) (string, error) {
	wrapped := p.wrap(body, at)
	return p.insert(func(node Producer) (string, error) {
		return node.At(wrapped, at)
	})
}

// Delay 延迟执行
func (p *producerCluster) Delay(body []byte, delay time.Duration) (string, error) {
	wrapped := p.wrap(body, time.Now().Add(delay))
	return p.insert(func(node Producer) (string, error) {
		return node.Delay(wrapped, delay)
	})
}

// Revoke 撤回一批任务
func (p *producerCluster) Revoke(ids string) error {
	var errs errorx.Errors

	fx.From(func(source chan<- interface{}) {
		for _, node := range p.nodes {
			source <- node
		}
	}).Map(func(item interface{}) interface{} {
		node := item.(Producer)
		return node.Revoke(ids)
	}).ForEach(func(item interface{}) {
		if item != nil {
			errs.Add(item.(error))
		}
	})

	return errs.Error()
}

// Close 关闭集群所有节点
func (p *producerCluster) Close() error {
	var errs errorx.Errors

	for _, node := range p.nodes {
		if err := node.Close(); err != nil {
			errs.Add(err)
		}
	}

	return errs.Error()
}

// 插入待处理任务
func (p *producerCluster) insert(fn func(node Producer) (string, error)) (string, error) {
	type idErr struct {
		id  string
		err error
	}
	var ret []idErr

	fx.From(func(source chan<- interface{}) {
		for _, node := range p.getWriteNodes() {
			source <- node
		}
	}).Map(func(item interface{}) interface{} {
		node := item.(Producer)
		id, err := fn(node)
		return idErr{
			id:  id,
			err: err,
		}
	}).ForEach(func(item interface{}) {
		ret = append(ret, item.(idErr))
	})

	var ids []string
	var errs errorx.Errors
	for _, val := range ret {
		if val.err != nil {
			errs.Add(val.err)
		} else {
			ids = append(ids, val.id)
		}
	}

	jointId := strings.Join(ids, idSep)
	if len(ids) >= minWrittenNodes {
		return jointId, nil
	}

	if err := p.Revoke(jointId); err != nil {
		logx.Error(err)
	}

	return "", errs.Error()
}

// wrap 将内容和执行时间包装为：UnixNano时间/内容
func (p *producerCluster) wrap(body []byte, at time.Time) []byte {
	var b bytes.Buffer
	b.WriteString(strconv.FormatInt(at.UnixNano(), 10))
	b.WriteByte(timeSep)
	b.Write(body)
	return b.Bytes()
}

// 获取可写节点
func (p *producerCluster) getWriteNodes() []Producer {
	if len(p.nodes) <= replicaNodes {
		return p.nodes
	}

	// 实际节点数比预设的多，则按预设数量随机选择
	nodes := p.cloneNodes()
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	return nodes[:replicaNodes]
}

// 克隆节点
func (p *producerCluster) cloneNodes() []Producer {
	return append([]Producer(nil), p.nodes...)
}
