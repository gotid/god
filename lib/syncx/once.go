package syncx

import "sync"

// Once 返回一个确保 fn 只会被调用一次的函数。
func Once(fn func()) func() {
	once := new(sync.Once)
	return func() {
		once.Do(fn)
	}
}
