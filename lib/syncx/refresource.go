package syncx

import (
	"errors"
	"sync"
)

// ErrUseOfCleaned 代表使用了已清理资源的错误。
var ErrUseOfCleaned = errors.New("使用了已清理的资源")

// RefResource 用于资源的引用计数。
type RefResource struct {
	lock    sync.Mutex
	ref     int32
	cleaned bool
	clean   func()
}

// NewRefResource 返回一个 RefResource。
func NewRefResource(clean func()) *RefResource {
	return &RefResource{
		clean: clean,
	}
}

// Use 使用资源时，递增资源的引用计数。
func (r *RefResource) Use() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.cleaned {
		// 此时引用已归零
		return ErrUseOfCleaned
	}

	r.ref++
	return nil
}

// Clean 清理资源时，递减资源的引用计数。
func (r *RefResource) Clean() {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.cleaned {
		return
	}

	r.ref--
	if r.ref == 0 {
		r.cleaned = true
		r.clean()
	}
}
