package syncx

import "sync"

// ManagedResource 托管资源。用于管理可能被破坏或重新获取的资源，例如连接。
type ManagedResource struct {
	resource interface{}
	lock     sync.RWMutex
	generate func() interface{}
	equal    func(a, b interface{}) bool
}

// NewManagedResource 返回一个托管资源。
func NewManagedResource(generate func() interface{}, equal func(a, b interface{}) bool) *ManagedResource {
	return &ManagedResource{
		generate: generate,
		equal:    equal,
	}
}

// MarkBroken 标记资源已受损。
func (mr *ManagedResource) MarkBroken(resource interface{}) {
	mr.lock.Lock()
	defer mr.lock.Unlock()

	if mr.equal(mr.resource, resource) {
		mr.resource = nil
	}
}

// Take 获取资源，有则返回，无则生成。
func (mr *ManagedResource) Take() interface{} {
	mr.lock.RLock()
	resource := mr.resource
	mr.lock.RUnlock()

	if resource != nil {
		return resource
	}

	mr.lock.Lock()
	defer mr.lock.Unlock()
	// 可能另一个 Take() 调用已经生成资源
	if mr.resource == nil {
		mr.resource = mr.generate()
	}
	return mr.resource
}
