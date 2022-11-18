package collection

import (
	"container/list"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mathx"
	"github.com/gotid/god/lib/syncx"
	"sync"
	"sync/atomic"
	"time"
)

const (
	statInterval     = 1 * time.Minute
	expiryDeviation  = 0.05
	defaultCacheName = "proc"
	slots            = 300
)

var emptyLruCache = emptyLru{}

type (
	// Cache 是一个基于内存的 LRU 缓存。
	Cache struct {
		name           string
		lock           sync.Mutex
		data           map[string]any
		expire         time.Duration
		timingWheel    *TimingWheel
		lruCache       lru
		barrier        syncx.SingleFlight
		unstableExpiry mathx.Unstable
		stats          *cacheStat
	}

	// CacheOption 定义自定义 Cache 选项的函数。
	CacheOption func(cache *Cache)
)

// NewCache 返回一个给定过期时长的 LRU Cache。
func NewCache(expire time.Duration, opts ...CacheOption) (*Cache, error) {
	cache := &Cache{
		data:           make(map[string]any),
		expire:         expire,
		lruCache:       emptyLruCache,
		barrier:        syncx.NewSingleFlight(),
		unstableExpiry: mathx.NewUnstable(expiryDeviation),
	}

	for _, opt := range opts {
		opt(cache)
	}

	if len(cache.name) == 0 {
		cache.name = defaultCacheName
	}
	cache.stats = newCacheStat(cache.name, cache.size)

	timingWheel, err := NewTimingWheel(time.Second, slots, func(key, val any) {
		k, ok := key.(string)
		if !ok {
			return
		}

		cache.Del(k)
	})
	if err != nil {
		return nil, err
	}

	cache.timingWheel = timingWheel
	return cache, nil
}

// Del 删除给定键对应的缓存。
func (c *Cache) Del(key string) {
	c.lock.Lock()
	delete(c.data, key)
	c.lruCache.remove(key)
	c.lock.Unlock()
	c.timingWheel.RemoveTimer(key)
}

// Get 获取给定键的缓存。
func (c *Cache) Get(key string) (any, bool) {
	value, ok := c.doGet(key)
	if ok {
		c.stats.IncrHit()
	} else {
		c.stats.IncrMiss()
	}

	return value, ok
}

// Set 设置键值对至缓存。
func (c *Cache) Set(key string, value any) {
	c.SetWithExpire(key, value, c.expire)
}

// SetWithExpire 设置给定存活时长的键值对至缓存。
func (c *Cache) SetWithExpire(key string, value any, expire time.Duration) {
	c.lock.Lock()
	_, ok := c.data[key]
	c.data[key] = value
	c.lruCache.add(key)
	c.lock.Unlock()

	expiry := c.unstableExpiry.AroundDuration(expire)
	if ok {
		c.timingWheel.MoveTimer(key, expiry)
	} else {
		c.timingWheel.SetTimer(key, value, expiry)
	}
}

// Take 返回给定键的条目。
// 如果条目在缓存中，则直接返回。
// 如果条目不在缓存中，使用获取方法获取后加入缓存并返回。
func (c *Cache) Take(key string, fetch func() (any, error)) (any, error) {
	if val, ok := c.doGet(key); ok {
		c.stats.IncrHit()
		return val, nil
	}

	var fresh bool
	val, err := c.barrier.Do(key, func() (any, error) {
		if val, ok := c.doGet(key); ok {
			return val, nil
		}

		v, e := fetch()
		if e != nil {
			return nil, e
		}

		fresh = true
		c.Set(key, v)

		return v, nil
	})
	if err != nil {
		return nil, err
	}

	if fresh {
		c.stats.IncrMiss()
		return val, nil
	}

	c.stats.IncrHit()
	return val, nil
}

func (c *Cache) size() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return len(c.data)
}

func (c *Cache) doGet(key string) (any, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	value, ok := c.data[key]
	if ok {
		c.lruCache.add(key)
	}

	return value, ok
}

func (c *Cache) onEvict(key string) {
	delete(c.data, key)
	c.timingWheel.RemoveTimer(key)
}

// WithLimit 限制缓存数量，符合 LRU 原则。
func WithLimit(limit int) CacheOption {
	return func(cache *Cache) {
		if limit > 0 {
			cache.lruCache = newKeyLru(limit, cache.onEvict)
		}
	}
}

// WithName 自定义缓存名称。
func WithName(name string) CacheOption {
	return func(cache *Cache) {
		cache.name = name
	}
}

type (
	lru interface {
		add(key string)
		remove(key string)
	}

	emptyLru struct{}

	keyLru struct {
		limit    int
		evicts   *list.List
		elements map[string]*list.Element
		onEvict  func(key string)
	}
)

func (e emptyLru) add(string)    {}
func (e emptyLru) remove(string) {}

func newKeyLru(limit int, onEvict func(key string)) *keyLru {
	return &keyLru{
		limit:    limit,
		evicts:   list.New(),
		elements: make(map[string]*list.Element),
		onEvict:  onEvict,
	}
}

func (k keyLru) add(key string) {
	if elem, ok := k.elements[key]; ok {
		k.evicts.MoveToFront(elem)
		return
	}

	// 添加新项
	elem := k.evicts.PushFront(key)
	k.elements[key] = elem

	// 验证是否越界
	if k.evicts.Len() > k.limit {
		k.removeOldest()
	}
}

func (k keyLru) remove(key string) {
	if elem, ok := k.elements[key]; ok {
		k.removeElement(elem)
	}
}

func (k keyLru) removeOldest() {
	elem := k.evicts.Back()
	if elem != nil {
		k.removeElement(elem)
	}
}

func (k keyLru) removeElement(elem *list.Element) {
	k.evicts.Remove(elem)
	key := elem.Value.(string)
	delete(k.elements, key)
	k.onEvict(key)
}

type cacheStat struct {
	name         string
	hit          uint64
	miss         uint64
	sizeCallback func() int
}

func newCacheStat(name string, sizeCallback func() int) *cacheStat {
	st := &cacheStat{
		name:         name,
		sizeCallback: sizeCallback,
	}
	go st.statLoop()
	return st
}

func (cs *cacheStat) IncrHit() {
	atomic.AddUint64(&cs.hit, 1)
}

func (cs *cacheStat) IncrMiss() {
	atomic.AddUint64(&cs.miss, 1)
}

func (cs *cacheStat) statLoop() {
	ticker := time.NewTicker(statInterval)
	defer ticker.Stop()

	for range ticker.C {
		hit := atomic.SwapUint64(&cs.hit, 0)
		miss := atomic.SwapUint64(&cs.miss, 0)
		total := hit + miss
		if total == 0 {
			continue
		}

		percent := 100 * float32(hit) / float32(total)
		logx.Statf("缓存(%s) - 请求数(m): %d, 命中率: %.1f%%, 元素: %d, 命中: %d, 未命中: %d",
			cs.name, total, percent, cs.sizeCallback(), hit, miss)
	}
}
