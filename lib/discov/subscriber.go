package discov

import (
	"github.com/gotid/god/lib/discov/internal"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"sync"
	"sync/atomic"
)

type (
	// SubOption 自定义订阅器的方法。
	SubOption func(sub *Subscriber)

	// Subscriber 用于订阅 etcd 集群中的给定键。
	Subscriber struct {
		endpoints []string
		exclusive bool
		items     *container
	}
)

// NewSubscriber 返回一个订阅器。
// endpoints 是 etcd 集群的主机列表。
// key 是待订阅的键。
// opts 用于自定义订阅器。
func NewSubscriber(endpoints []string, key string, opts ...SubOption) (*Subscriber, error) {
	sub := &Subscriber{
		endpoints: endpoints,
	}
	for _, opt := range opts {
		opt(sub)
	}
	sub.items = newContainer(sub.exclusive)

	if err := internal.GetRegistry().Monitor(endpoints, key, sub.items); err != nil {
		return nil, err
	}

	return sub, nil
}

// AddListener 添加订阅器的监听器。
func (s *Subscriber) AddListener(listener func()) {
	s.items.addListener(listener)
}

// Values 返回所有订阅值。
func (s *Subscriber) Values() []string {
	return s.items.getValues()
}

// Exclusive 意为键值必须1比1，也就是后续关联的值会替换之前的值。
func Exclusive() SubOption {
	return func(sub *Subscriber) {
		sub.exclusive = true
	}
}

// WithSubEtcdAccount 提供 etcd 用户名/密码。
func WithSubEtcdAccount(user, pass string) SubOption {
	return func(sub *Subscriber) {
		RegisterAccount(sub.endpoints, user, pass)
	}
}

// WithSubEtcdTLS 提供 etcd CertFile/CertKeyFile/CACertFile.
func WithSubEtcdTLS(certFile, certKeyFile, caFile string, insecureSkipVerify bool) SubOption {
	return func(sub *Subscriber) {
		logx.Must(RegisterTLS(sub.endpoints, certFile, certKeyFile, caFile, insecureSkipVerify))
	}
}

type container struct {
	exclusive bool
	values    map[string][]string
	mapping   map[string]string
	snapshot  atomic.Value
	dirty     *syncx.AtomicBool
	listeners []func()
	lock      sync.Mutex
}

func newContainer(exclusive bool) *container {
	return &container{
		exclusive: exclusive,
		values:    make(map[string][]string),
		mapping:   make(map[string]string),
		dirty:     syncx.ForAtomicBool(true),
	}
}

func (c *container) OnAdd(kv internal.KV) {
	c.addKv(kv.Key, kv.Val)
	c.notifyChange()
}

func (c *container) OnDelete(kv internal.KV) {
	c.removeKey(kv.Key)
	c.notifyChange()
}

// 添加键值对，如果有其他键与值关联，则返回
func (c *container) addKv(key string, val string) ([]string, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.dirty.Set(true)
	keys := c.values[val]
	previous := append([]string(nil), keys...)
	early := len(keys) > 0
	if c.exclusive && early {
		for _, k := range keys {
			c.doRemoveKey(k)
		}
	}
	c.values[val] = append(c.values[val], key)
	c.mapping[key] = val

	if early {
		return previous, true
	}

	return nil, false
}

func (c *container) removeKey(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.dirty.Set(true)
	c.doRemoveKey(key)
}

func (c *container) doRemoveKey(key string) {
	server, ok := c.mapping[key]
	if !ok {
		return
	}

	delete(c.mapping, key)
	keys := c.values[server]
	remain := keys[:0]

	for _, k := range keys {
		if k != key {
			remain = append(remain, k)
		}
	}

	if len(remain) > 0 {
		c.values[server] = remain
	} else {
		delete(c.values, server)
	}
}

func (c *container) notifyChange() {
	c.lock.Lock()
	listeners := append(([]func())(nil), c.listeners...)
	c.lock.Unlock()

	for _, listener := range listeners {
		listener()
	}
}

func (c *container) addListener(listener func()) {
	c.lock.Lock()
	c.listeners = append(c.listeners, listener)
	c.lock.Unlock()
}

func (c *container) getValues() []string {
	if !c.dirty.True() {
		return c.snapshot.Load().([]string)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	var vals []string
	for each := range c.values {
		vals = append(vals, each)
	}
	c.snapshot.Store(vals)
	c.dirty.Set(false)

	return vals
}
