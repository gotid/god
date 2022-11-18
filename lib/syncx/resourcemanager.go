package syncx

import (
	"github.com/gotid/god/lib/errorx"
	"io"
	"sync"
)

// ResourceManager 是一个用于管理资源的管理器。
type ResourceManager struct {
	resources    map[string]io.Closer
	singleFlight SingleFlight
	lock         sync.RWMutex
}

// NewResourceManager 返回一个 ResourceManager。
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources:    make(map[string]io.Closer),
		singleFlight: NewSingleFlight(),
	}
}

// Close 关闭管理器。
// 在调用 Close 之后不要在使用 ResourceManager。
func (m *ResourceManager) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var be errorx.BatchError
	for _, resource := range m.resources {
		if err := resource.Close(); err != nil {
			be.Add(err)
		}
	}

	m.resources = nil

	return be.Err()
}

// Get 返回给定键的资源。
func (m *ResourceManager) Get(key string, create func() (io.Closer, error)) (io.Closer, error) {
	val, err := m.singleFlight.Do(key, func() (any, error) {
		m.lock.RLock()
		resource, ok := m.resources[key]
		m.lock.RUnlock()
		if ok {
			return resource, nil
		}

		resource, err := create()
		if err != nil {
			return nil, err
		}

		m.lock.Lock()
		defer m.lock.Unlock()
		m.resources[key] = resource

		return resource, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(io.Closer), nil
}

// Set 设置给定键的资源。
func (m *ResourceManager) Set(key string, resource io.Closer) {
	m.lock.Lock()
	m.resources[key] = resource
	m.lock.Unlock()
}
