package cache

import (
	"fmt"
	"time"

	"github.com/gotid/god/lib/errorx"
	"github.com/gotid/god/lib/hash"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
)

type (
	cluster struct {
		dispatcher  *hash.ConsistentHash
		errNotFound error
	}
)

func New(conf ClusterConf, barrier syncx.SingleFlight, stat *Stat,
	errNotFound error, opts ...Option) Cache {
	if len(conf) == 0 || TotalWeights(conf) <= 0 {
		logx.Error("未配置缓存节点")
	}

	if len(conf) == 1 {
		return NewNode(conf[0].NewRedis(), barrier, stat, errNotFound, opts...)
	}

	// 添加一批 redis 缓存节点
	dispatcher := hash.NewConsistentHash()
	for _, conf := range conf {
		node := NewNode(conf.NewRedis(), barrier, stat, errNotFound, opts...)
		dispatcher.AddWithWeight(node, conf.Weight)
	}

	return cluster{
		dispatcher:  dispatcher,
		errNotFound: errNotFound,
	}
}

func (c cluster) Del(keys ...string) error {
	switch len(keys) {
	case 0:
		return nil
	case 1:
		key := keys[0]
		node, ok := c.dispatcher.Get(key)
		if !ok {
			return c.errNotFound
		}
		return node.(Cache).Del(key)
	default:
		var es errorx.Errors
		nodes := make(map[interface{}][]string)
		for _, key := range keys {
			node, ok := c.dispatcher.Get(key)
			if !ok {
				es.Add(fmt.Errorf("缓存 key %q 不存在", key))
				continue
			}

			nodes[node] = append(nodes[node], key)
		}
		for node, keys := range nodes {
			if err := node.(Cache).Del(keys...); err != nil {
				es.Add(err)
			}
		}

		return es.Error()
	}
}

func (c cluster) Get(key string, dest interface{}) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).Get(key, dest)
}

func (c cluster) MGet(keys []string, dest []interface{}) error {
	switch len(keys) {
	case 0:
		return nil
	case 1:
		key := keys[0]
		node, ok := c.dispatcher.Get(key)
		if !ok {
			return c.errNotFound
		}
		return node.(Cache).MGet(keys, dest)
	default:
		var es errorx.Errors
		nodes := make(map[interface{}][]string)
		for _, key := range keys {
			node, ok := c.dispatcher.Get(key)
			if !ok {
				es.Add(fmt.Errorf("缓存 key %q 不存在", key))
				continue
			}

			nodes[node] = append(nodes[node], key)
		}

		for node, keys := range nodes {
			err := node.(Cache).MGet(keys, dest)
			if err != nil {
				es.Add(err)
			}
		}

		return es.Error()
	}
}

func (c cluster) Set(key string, value interface{}) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).Set(key, value)
}

func (c cluster) SetEx(key string, value interface{}, expires time.Duration) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).SetEx(key, value, expires)
}

func (c cluster) SetBit(key string, offset int64, value int) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).SetBit(key, offset, value)
}

func (c cluster) SetBits(key string, offset []int64) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).SetBits(key, offset)
}

func (c cluster) UnsetBits(key string, offset []int64) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).UnsetBits(key, offset)
}

func (c cluster) GetBits(key string, offset []int64) (map[int64]bool, error) {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return nil, c.errNotFound
	}

	return node.(Cache).GetBits(key, offset)
}

func (c cluster) GetBit(key string, offset int64) (int, error) {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return 0, c.errNotFound
	}

	return node.(Cache).GetBit(key, offset)
}

func (c cluster) Take(dest interface{}, key string, queryFn func(v interface{}) error) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).Take(dest, key, queryFn)
}

func (c cluster) TakeEx(dest interface{}, key string, queryFn func(newVal interface{}, expires time.Duration) error) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).TakeEx(dest, key, queryFn)
}
