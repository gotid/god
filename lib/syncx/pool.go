package syncx

import (
	"github.com/gotid/god/lib/timex"
	"sync"
	"time"
)

type (
	// Pool 用于存取临时对象的资源池。
	// 与 sync.Pool 的区别是：
	// 1. 限制资源的数量
	// 2. 可以设置资源的最大使用期限
	// 3. 可以自定义资源的销毁方法
	Pool struct {
		limit   int
		created int
		maxAge  time.Duration
		lock    sync.Locker
		cond    *sync.Cond
		head    *node
		create  func() any
		destroy func(any)
	}

	// PoolOption 自定义 Pool 的方法。
	PoolOption func(*Pool)

	node struct {
		item     any
		next     *node
		lastUsed time.Duration
	}
)

// NewPool 返回一个 Pool。
//
// n 为资源池大小，create 为资源不存在的创建函数，destroy 为资源销毁函数。
func NewPool(n int, create func() any, destroy func(any), opts ...PoolOption) *Pool {
	if n <= 0 {
		panic("池大小不能为负数或零")
	}

	lock := new(sync.Mutex)
	pool := &Pool{
		limit:   n,
		lock:    lock,
		cond:    sync.NewCond(lock),
		create:  create,
		destroy: destroy,
	}

	for _, opt := range opts {
		opt(pool)
	}

	return pool
}

// Get 从池中取出一个资源。
func (p *Pool) Get() any {
	p.lock.Lock()
	defer p.lock.Unlock()

	for {
		if p.head != nil {
			head := p.head
			p.head = head.next
			if p.maxAge > 0 && head.lastUsed+p.maxAge < timex.Now() {
				p.created--
				p.destroy(head.item)
				continue
			} else {
				return head.item
			}
		}

		if p.created < p.limit {
			p.created++
			return p.create()
		}

		p.cond.Wait()
	}
}

// Put 放入一个资源到池中。
func (p *Pool) Put(x any) {
	if x == nil {
		return
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	p.head = &node{
		item:     x,
		next:     p.head,
		lastUsed: timex.Now(),
	}
	p.cond.Signal()
}

// WithMaxAge 返回一个自定义 Pool 使用期限的函数。
func WithMaxAge(duration time.Duration) PoolOption {
	return func(pool *Pool) {
		pool.maxAge = duration
	}
}
