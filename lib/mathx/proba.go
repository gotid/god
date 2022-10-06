package mathx

import (
	"math/rand"
	"sync"
	"time"
)

// Proba 用于测试给定的可能性是否为真。
type Proba struct {
	// rand.New(...) 返回非线程安全对象
	r    *rand.Rand
	lock sync.Mutex
}

// NewProba 返回一个 Proba。
func NewProba() *Proba {
	return &Proba{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// TrueOnProba 判断给定的可能性是否为真。
func (p *Proba) TrueOnProba(proba float64) (truth bool) {
	p.lock.Lock()
	truth = p.r.Float64() < proba
	p.lock.Unlock()
	return
}
