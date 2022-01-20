package syncx

import (
	"io"
	"sync"

	"github.com/gotid/god/lib/errorx"
)

// ResourceManager 是一个管理复用资源的管理器。
type ResourceManager struct {
	resources    map[string]io.Closer
	singleFlight SingleFlight
	lock         sync.RWMutex
}

// NewResourceManager 返回一个复用资源管理器。
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources:    make(map[string]io.Closer),
		singleFlight: NewSingleFlight(),
	}
}

// Close 关闭该复用资源管理器。
func (m *ResourceManager) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var errs errorx.Errors
	for _, r := range m.resources {
		if err := r.Close(); err != nil {
			errs.Add(err)
		}
	}

	// 释放资源
	m.resources = nil

	return errs.Error()
}

// Get 获取指定 key 的缓存资源。如不存在，则调用创建函数回源获取并缓存。
func (m *ResourceManager) Get(key string, create func() (io.Closer, error)) (io.Closer, error) {
	val, _, err := m.singleFlight.Do(key, func() (interface{}, error) {
		m.lock.Lock()
		res, ok := m.resources[key]
		m.lock.Unlock()
		if ok {
			return res, nil
		}

		res, err := create()
		if err != nil {
			return nil, err
		}

		m.lock.Lock()
		m.resources[key] = res
		m.lock.Unlock()

		return res, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(io.Closer), nil
}
