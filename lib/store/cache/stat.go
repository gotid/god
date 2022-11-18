package cache

import (
	"github.com/gotid/god/lib/logx"
	"sync/atomic"
	"time"
)

const statInterval = time.Minute

// Stat 用于统计缓存状态。
type Stat struct {
	name    string
	Total   uint64
	Hit     uint64
	Miss    uint64
	DbFails uint64
}

// NewStat 返回一个给定名称的缓存统计。
func NewStat(name string) *Stat {
	s := &Stat{
		name: name,
	}
	go s.statLoop()

	return s
}

// IncrTotal 递增请求总次数。
func (s *Stat) IncrTotal() {
	atomic.AddUint64(&s.Total, 1)
}

// IncrHit 递增击中次数。
func (s *Stat) IncrHit() {
	atomic.AddUint64(&s.Hit, 1)
}

// IncrMiss 递增未击中次数。
func (s *Stat) IncrMiss() {
	atomic.AddUint64(&s.Miss, 1)
}

// IncrDbFails 递增数据库失败次数。
func (s *Stat) IncrDbFails() {
	atomic.AddUint64(&s.DbFails, 1)
}

func (s *Stat) statLoop() {
	ticker := time.NewTicker(statInterval)
	defer ticker.Stop()

	for range ticker.C {
		total := atomic.SwapUint64(&s.Total, 0)
		if total == 0 {
			continue
		}

		hit := atomic.SwapUint64(&s.Hit, 0)
		percent := 100 * float32(hit) / float32(total)
		miss := atomic.SwapUint64(&s.Miss, 0)
		dbFails := atomic.SwapUint64(&s.DbFails, 0)
		logx.Statf("数据库缓存(%s) - 请求数(m): %d, 命中率: %.1f%%, 命中: %d, 未命中: %d, 数据库错误L %d",
			s.name, total, percent, hit, miss, dbFails)
	}
}
