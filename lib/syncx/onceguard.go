package syncx

import "sync/atomic"

// OnceGuard 用于确保资源只能被获取一次。
type OnceGuard struct {
	done uint32
}

// Take 获取资源，成功返回true，失败返回false。
func (og *OnceGuard) Take() bool {
	return atomic.CompareAndSwapUint32(&og.done, 0, 1)
}

// Taken 判断资源是否已被获取。
func (og *OnceGuard) Taken() bool {
	return atomic.LoadUint32(&og.done) == 1
}
