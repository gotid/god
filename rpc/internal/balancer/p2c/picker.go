package p2c

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
	"github.com/gotid/god/rpc/internal/codes"
	"google.golang.org/grpc/balancer"
)

type p2cPicker struct {
	conns []*subConn
	r     *rand.Rand
	stamp *syncx.AtomicDuration
	lock  sync.Mutex
}

func (p *p2cPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	var chosen *subConn
	switch len(p.conns) {
	case 0: // 没有连接
		return balancer.PickResult{SubConn: nil, Done: nil}, balancer.ErrNoSubConnAvailable
	case 1: // 一个连接
		chosen = p.choose(p.conns[0], nil)
	case 2: // 2个连接
		chosen = p.choose(p.conns[0], p.conns[1])
	default: // 2个以上连接，随机选举两个健康的并在其中选举1个
		var c1, c2 *subConn
		for i := 0; i < pickTimes; i++ {
			a := p.r.Intn(len(p.conns))
			b := p.r.Intn(len(p.conns) - 1)
			if b >= a {
				b++
			}
			c1 = p.conns[a]
			c2 = p.conns[b]
			if c1.healthy() && c2.healthy() {
				break
			}
		}

		chosen = p.choose(c1, c2)
	}

	atomic.AddInt64(&chosen.inflight, 1) // 飞行中+1
	atomic.AddInt64(&chosen.requests, 1) // 请求数+1

	return balancer.PickResult{
		SubConn: chosen.conn,
		Done:    p.buildDoneFunc(chosen),
	}, nil
}

// choose 从两个连接中选举一个用于使用
func (p *p2cPicker) choose(c1, c2 *subConn) *subConn {
	start := int64(timex.Now())
	if c2 == nil {
		atomic.StoreInt64(&c1.pick, start)
		return c1
	}

	if c1.load() > c2.load() {
		c1, c2 = c2, c1
	}

	pick := atomic.LoadInt64(&c2.pick)
	if start-pick > forcePickTime && atomic.CompareAndSwapInt64(&c2.pick, pick, start) {
		return c2
	}

	atomic.StoreInt64(&c1.pick, start)
	return c1
}

// buildDoneFunc 构建完成连接函数
func (p *p2cPicker) buildDoneFunc(conn *subConn) func(doneInfo balancer.DoneInfo) {
	start := int64(timex.Now())
	return func(info balancer.DoneInfo) {
		atomic.AddInt64(&conn.inflight, -1)
		now := timex.Now()
		last := atomic.SwapInt64(&conn.last, int64(now))
		duration := int64(now) - last
		if duration < 0 {
			duration = 0
		}
		w := math.Exp(float64(-duration) / float64(decayTime))
		lag := int64(now) - start
		if lag < 0 {
			lag = 0
		}
		oldLag := atomic.LoadUint64(&conn.lag)
		if oldLag == 0 {
			w = 0
		}
		atomic.StoreUint64(&conn.lag, uint64(float64(oldLag)*w+float64(lag)*(1-w)))
		success := initSuccess
		if info.Err != nil && !codes.Acceptable(info.Err) {
			success = 0
		}
		oldSuccess := atomic.LoadUint64(&conn.success)
		atomic.StoreUint64(&conn.success, uint64(float64(oldSuccess)*w+float64(success)*(1-w)))

		stamp := p.stamp.Load()
		if now-stamp >= logInterval {
			if p.stamp.CompareAndSwap(stamp, now) {
				p.logStats()
			}
		}
	}
}

func (p *p2cPicker) logStats() {
	var stats []string

	p.lock.Lock()
	defer p.lock.Unlock()

	for _, conn := range p.conns {
		stats = append(stats, fmt.Sprintf("conn: %s, load: %d, reqs: %d",
			conn.addr.Addr, conn.load(), atomic.SwapInt64(&conn.requests, 0)))
	}

	logx.Statf("p2c - %s", strings.Join(stats, "; "))
}
