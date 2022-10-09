package collection

import "sync"

const (
	maxDeletion   = 10000
	copyThreshold = 1000
)

// SafeMap 避免内存泄露的 map 版本。
type SafeMap struct {
	lock        sync.RWMutex
	deletionOld int
	deletionNew int
	dirtyOld    map[interface{}]interface{}
	dirtyNew    map[interface{}]interface{}
}

// NewSafeMap 返回一个 SafeMap。
func NewSafeMap() *SafeMap {
	return &SafeMap{
		dirtyOld: make(map[interface{}]interface{}),
		dirtyNew: make(map[interface{}]interface{}),
	}
}

// Del 删除 m 中指定键值。
func (m *SafeMap) Del(key interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.dirtyOld[key]; ok {
		delete(m.dirtyOld, key)
		m.deletionOld++
	} else if _, ok := m.dirtyNew[key]; ok {
		delete(m.dirtyNew, key)
		m.deletionNew++
	}

	if m.deletionOld >= maxDeletion && len(m.dirtyOld) < copyThreshold {
		for k, v := range m.dirtyOld {
			m.dirtyNew[k] = v
		}
		m.dirtyOld = m.dirtyNew
		m.deletionOld = m.deletionNew
		m.dirtyNew = make(map[interface{}]interface{})
		m.deletionNew = 0
	}
	if m.deletionNew >= maxDeletion && len(m.dirtyNew) < copyThreshold {
		for k, v := range m.dirtyNew {
			m.dirtyOld[k] = v
		}
		m.dirtyNew = make(map[interface{}]interface{})
		m.deletionNew = 0
	}
}

// Get 获取 m 中给定键的值。
func (m *SafeMap) Get(key interface{}) (interface{}, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if val, ok := m.dirtyOld[key]; ok {
		return val, true
	}

	val, ok := m.dirtyNew[key]
	return val, ok
}

// Put 将指定键值对加入 m。
func (m *SafeMap) Put(key, val interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.deletionOld <= maxDeletion {
		if _, ok := m.dirtyNew[key]; ok {
			delete(m.dirtyNew, key)
			m.deletionNew++
		}
		m.dirtyOld[key] = val
	} else {
		if _, ok := m.dirtyOld[key]; ok {
			delete(m.dirtyOld, key)
			m.deletionOld++
		}
		m.dirtyNew[key] = val
	}
}

// Size 返回 m 的大小。
func (m *SafeMap) Size() int {
	m.lock.RLock()
	size := len(m.dirtyOld) + len(m.dirtyNew)
	m.lock.RUnlock()
	return size
}

// Range 让 m 中的每个成员都调用 f 函数，如果返回 false 则退出迭代。
func (m *SafeMap) Range(f func(key, val interface{}) bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for k, v := range m.dirtyOld {
		if !f(k, v) {
			return
		}
	}

	for k, v := range m.dirtyNew {
		if !f(k, v) {
			return
		}
	}
}
