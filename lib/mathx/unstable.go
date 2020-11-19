package mathx

import (
	"math/rand"
	"sync"
	"time"
)

// Unstable 值会在[0, 1]之间上下浮动的偏差结构体
type Unstable struct {
	deviation float64
	r         *rand.Rand
	lock      *sync.Mutex
}

// NewUnstable 新建一个值会浮动的偏差结构体
func NewUnstable(deviation float64) Unstable {
	if deviation < 0 {
		deviation = 0
	}
	if deviation > 1 {
		deviation = 1
	}
	return Unstable{
		deviation: deviation,
		r:         rand.New(rand.NewSource(time.Now().UnixNano())),
		lock:      new(sync.Mutex),
	}
}

// AroundDuration 基于指定时间段，生成临近的不同的值
func (u Unstable) AroundDuration(base time.Duration) time.Duration {
	u.lock.Lock()
	val := time.Duration((1 + u.deviation - 2*u.deviation*u.r.Float64()) * float64(base))
	u.lock.Unlock()
	return val
}

func (u Unstable) AroundInt(base int64) int64 {
	u.lock.Lock()
	val := int64((1 + u.deviation - 2*u.deviation*u.r.Float64()) * float64(base))
	u.lock.Unlock()
	return val
}
