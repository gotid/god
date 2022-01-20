package collection

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mathx"
	"github.com/gotid/god/lib/syncx"
)

const (
	defaultCacheName = "proc" // 进程内缓存名
	slots            = 300
	statInterval     = time.Minute // 缓存统计间隔时长，默认为1分钟

	// 缓存过期偏差值：通过过期偏差，避免大量缓存同时过期
	// 将缓存到期时间偏差上下设置为0.05，也就是落到[0.95,1.05]*秒
	expireDeviation = 0.05
)

var emptyLruCache = emptyLru{}

type (
	// Cache 表示一个内存中的缓存对象(in-memory)。
	Cache struct {
		name           string
		lock           sync.Mutex
		data           map[string]interface{}
		expire         time.Duration
		timingWheel    *TimingWheel
		lruCache       lru
		barrier        syncx.SingleFlight
		unstableExpiry mathx.Unstable
		stats          *cacheStat
	}

	// CacheOption 表示一个自定义 Cache 的函数。
	CacheOption func(cache *Cache)
)

// NewCache 新建一个进程内缓存
func NewCache(expire time.Duration, opts ...CacheOption) (*Cache, error) {
	cache := &Cache{
		data:           make(map[string]interface{}),
		expire:         expire,
		lruCache:       emptyLruCache,
		barrier:        syncx.NewSingleFlight(),
		unstableExpiry: mathx.NewUnstable(expireDeviation),
	}

	for _, opt := range opts {
		opt(cache)
	}

	if len(cache.name) == 0 {
		cache.name = defaultCacheName
	}
	cache.stats = newCacheStat(cache.name, cache.size)

	timingWheel, err := NewTimingWheel(time.Second, slots, func(k, v interface{}) {
		key, ok := k.(string)
		if !ok {
			return
		}

		cache.Del(key)
	})
	if err != nil {
		return nil, err
	}

	cache.timingWheel = timingWheel
	return cache, nil
}

// WithLimit 自定义缓存条数的函数。
func WithLimit(limit int) CacheOption {
	return func(cache *Cache) {
		if limit > 0 {
			cache.lruCache = newKeyLru(limit, cache.onEvict)
		}
	}
}

// WithName 自定义缓存名称的函数。
func WithName(name string) CacheOption {
	return func(cache *Cache) {
		cache.name = name
	}
}

// Del 从缓存中删除指定键。
func (c *Cache) Del(key string) {
	c.lock.Lock()
	delete(c.data, key)
	c.lruCache.remove(key)
	c.lock.Unlock()
	c.timingWheel.RemoveTimer(key)
}

// Get 从缓存中获取指定键的值。
func (c *Cache) Get(key string) (interface{}, bool) {
	v, ok := c.doGet(key)
	if ok {
		c.stats.IncrHit()
	} else {
		c.stats.IncrMiss()
	}

	return v, ok
}

func (c *Cache) Set(key string, value interface{}) {
	c.lock.Lock()
	_, ok := c.data[key]
	c.data[key] = value
	c.lruCache.add(key)
	c.lock.Unlock()

	expiry := c.unstableExpiry.AroundDuration(c.expire)
	if ok {
		c.timingWheel.MoveTimer(key, expiry)
	} else {
		c.timingWheel.SetTimer(key, value, expiry)
	}
}

// Take 有缓存则返回，无则获取（通过共享调用实现高并发）
func (c *Cache) Take(key string, fetch func() (interface{}, error)) (interface{}, error) {
	if val, ok := c.doGet(key); ok {
		c.stats.IncrHit()
		return val, nil
	}

	// 回源获取，通过共享调用实现高并发
	val, hit, err := c.barrier.Do(key, func() (interface{}, error) {
		// 因为在内存中进行的map搜索，时间为O(1)，而fetch是在IO上的查询
		// 所以我们做双重检测，因为缓存有可能被其他调用取到了
		if val, ok := c.doGet(key); ok {
			return val, nil
		}

		v, e := fetch()
		if e != nil {
			return nil, e
		}

		c.Set(key, v)
		return v, nil
	})
	if err != nil {
		return nil, err
	}

	if hit {
		c.stats.IncrHit()
	} else {
		c.stats.IncrMiss()
	}

	return val, nil
}

func (c *Cache) doGet(key string) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.data[key]
	if ok {
		c.lruCache.add(key)
	}

	return val, ok
}

func (c *Cache) onEvict(key string) {
	delete(c.data, key)
	c.timingWheel.RemoveTimer(key)
}

func (c *Cache) size() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return len(c.data)
}

//// LRU
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

func newKeyLru(limit int, onEvict func(key string)) *keyLru {
	return &keyLru{
		limit:    limit,
		evicts:   list.New(),
		elements: make(map[string]*list.Element),
		onEvict:  onEvict,
	}
}

func (l emptyLru) add(string) {}

func (l emptyLru) remove(string) {}

func (kl *keyLru) add(key string) {
	if elem, ok := kl.elements[key]; ok {
		kl.evicts.MoveToFront(elem)
		return
	}

	// 添加新项
	elem := kl.evicts.PushFront(key)
	kl.elements[key] = elem

	// 防止下标溢出
	if kl.evicts.Len() > kl.limit {
		kl.removeOldest()
	}
}

func (kl *keyLru) remove(key string) {
	if elem, ok := kl.elements[key]; ok {
		kl.removeElement(elem)
	}
}

func (kl *keyLru) removeOldest() {
	elem := kl.evicts.Back()
	if elem != nil {
		kl.removeElement(elem)
	}
}

func (kl *keyLru) removeElement(e *list.Element) {
	kl.evicts.Remove(e)
	key := e.Value.(string)
	delete(kl.elements, key)
	kl.onEvict(key)
}

// cacheStat 表示一个缓存统计项。
type cacheStat struct {
	name         string
	hit          uint64
	miss         uint64
	sizeCallback func() int
}

func newCacheStat(name string, sizeCallback func() int) *cacheStat {
	s := &cacheStat{
		name:         name,
		sizeCallback: sizeCallback,
	}
	go s.loop()
	return s
}

func (s *cacheStat) loop() {
	ticker := time.NewTicker(statInterval)
	defer ticker.Stop()

	for range ticker.C {
		hit := atomic.SwapUint64(&s.hit, 0)
		miss := atomic.SwapUint64(&s.miss, 0)
		total := hit + miss
		if total == 0 {
			continue
		}
		hitRatio := 100 * float32(hit) / float32(total)
		logx.Statf("缓存(%s) - 一分钟请求数: %d, 命中率: %.1f%%, 成员: %d, 命中: %d, 未命中: %d",
			s.name, total, hitRatio, s.sizeCallback(), hit, miss)
	}
}

func (s *cacheStat) IncrHit() {
	atomic.AddUint64(&s.hit, 1)
}

func (s *cacheStat) IncrMiss() {
	atomic.AddUint64(&s.miss, 1)
}
