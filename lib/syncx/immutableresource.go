package syncx

import (
	"github.com/gotid/god/lib/timex"
	"sync"
	"time"
)

const defaultRefreshInterval = time.Second

type (
	// ImmutableResource 用于管理一个不可变资源。
	ImmutableResource struct {
		fetch           func() (interface{}, error)
		resource        interface{}
		err             error
		lock            sync.RWMutex
		refreshInterval time.Duration
		lastTime        *AtomicDuration
	}

	// ImmutableResourceOption 自定义 ImmutableResource 的方法。
	ImmutableResourceOption func(resource *ImmutableResource)
)

// NewImmutableResource 返回一个 ImmutableResource。
func NewImmutableResource(fetch func() (interface{}, error), opts ...ImmutableResourceOption) *ImmutableResource {
	ir := ImmutableResource{
		fetch:           fetch,
		refreshInterval: defaultRefreshInterval,
		lastTime:        NewAtomicDuration(),
	}
	for _, opt := range opts {
		opt(&ir)
	}
	return &ir
}

// Get 获取不可变资源，有资源直接返回，无资源尝试获取。
func (ir *ImmutableResource) Get() (interface{}, error) {
	ir.lock.RLock()
	resource := ir.resource
	ir.lock.RUnlock()
	if resource != nil {
		return resource, nil
	}

	ir.maybeRefresh(func() {
		res, err := ir.fetch()
		ir.lock.Lock()
		if err != nil {
			ir.err = err
		} else {
			ir.resource, ir.err = res, nil
		}
		ir.lock.Unlock()
	})

	ir.lock.RLock()
	resource, err := ir.resource, ir.err
	ir.lock.RUnlock()
	return resource, err
}

func (ir *ImmutableResource) maybeRefresh(execute func()) {
	now := timex.Now()
	lastTime := ir.lastTime.Load()
	if lastTime == 0 || lastTime+ir.refreshInterval < now {
		ir.lastTime.Set(now)
		execute()
	}
}

// WithRefreshIntervalOnFailure 设置失败时的刷新时间。
// interval 为 0 会在失败时强制刷新。默认单位为 time.Second。
func WithRefreshIntervalOnFailure(interval time.Duration) ImmutableResourceOption {
	return func(resource *ImmutableResource) {
		resource.refreshInterval = interval
	}
}
