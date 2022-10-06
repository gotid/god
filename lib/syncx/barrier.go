package syncx

import "sync"

// Barrier 用于在一个资源上设置并发屏障。
type Barrier struct {
	lock sync.Mutex
}

// Guard 用互斥锁保卫给定的函数 fn。
func (b *Barrier) Guard(fn func()) {
	Guard(&b.lock, fn)
}

// Guard 用锁保卫给定的函数 fn。
func Guard(lock sync.Locker, fn func()) {
	lock.Lock()
	defer lock.Unlock()
	fn()
}
