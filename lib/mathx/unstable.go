package mathx

import (
	"math/rand"
	"sync"
	"time"
)

// Unstable 用于根据给定偏差围绕平均值生成一个随机值。
type Unstable struct {
	deviation float64
	r         *rand.Rand
	lock      *sync.Mutex
}

// NewUnstable 返回一个不稳固值的实例。
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

// AroundDuration 根据给定的基准时长和公差生成一个随机周边时长，± u.deviation。
func (u Unstable) AroundDuration(base time.Duration) time.Duration {
	u.lock.Lock()
	val := time.Duration((1 + u.deviation - 2*u.deviation*u.r.Float64()) * float64(base))
	u.lock.Unlock()
	return val
}

// AroundInt 根据给定的基准数值和公差生成一个随机的周边数值，± u.deviation。
func (u Unstable) AroundInt(base int64) int64 {
	u.lock.Lock()
	val := int64((1 + u.deviation - 2*u.deviation*u.r.Float64()) * float64(base))
	u.lock.Unlock()
	return val
}
