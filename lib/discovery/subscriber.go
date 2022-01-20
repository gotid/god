package discovery

import (
	"sync"
	"sync/atomic"

	"github.com/gotid/god/lib/discovery/internal"
	"github.com/gotid/god/lib/syncx"
)

type (
	subOptions struct {
		exclude bool
	}

	SubOption func(opts *subOptions)

	container struct {
		exclude   bool
		values    map[string][]string
		mapping   map[string]string
		snapshot  atomic.Value
		dirty     *syncx.AtomicBool
		listeners []func()
		lock      sync.Mutex
	}

	Subscriber struct {
		items *container
	}
)

func NewSubscriber(endpoints []string, key string, opts ...SubOption) (*Subscriber, error) {
	var subOpts subOptions
	for _, opt := range opts {
		opt(&subOpts)
	}

	sub := &Subscriber{items: newContainer(subOpts.exclude)}
	if err := internal.GetRegistry().Monitor(endpoints, key, sub.items); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *Subscriber) AddListener(listener func()) {
	s.items.addListener(listener)
}

func (s *Subscriber) Values() []string {
	return s.items.getValues()
}

func Exclude() SubOption {
	return func(opts *subOptions) {
		opts.exclude = true
	}
}

func newContainer(exclude bool) *container {
	return &container{
		exclude: exclude,
		values:  make(map[string][]string),
		mapping: make(map[string]string),
		dirty:   syncx.ForAtomicBool(true),
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

// addKv adds the kv, returns if there are already other keys associate with the value
func (c *container) addKv(key, value string) ([]string, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.dirty.Set(true)
	keys := c.values[value]
	previous := append([]string(nil), keys...)
	early := len(keys) > 0
	if c.exclude && early {
		for _, each := range keys {
			c.doRemoveKey(each)
		}
	}
	c.values[value] = append(c.values[value], key)
	c.mapping[key] = value

	if early {
		return previous, true
	} else {
		return nil, false
	}
}

func (c *container) addListener(listener func()) {
	c.lock.Lock()
	c.listeners = append(c.listeners, listener)
	c.lock.Unlock()
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

func (c *container) notifyChange() {
	c.lock.Lock()
	listeners := append(([]func())(nil), c.listeners...)
	c.lock.Unlock()

	for _, listener := range listeners {
		listener()
	}
}

// removeKey removes the kv, returns true if there are still other keys associate with the value
func (c *container) removeKey(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.dirty.Set(true)
	c.doRemoveKey(key)
}
