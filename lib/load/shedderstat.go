package load

import (
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
)

type (
	// ShedderStat 泄流阀统计项。
	ShedderStat struct {
		name  string
		total int64
		pass  int64
		drop  int64
	}

	snapshot struct {
		Total int64
		Pass  int64
		Drop  int64
	}
)

// NewShedderStat 返回一个命名的泄流阀统计项。
func NewShedderStat(name string) *ShedderStat {
	s := &ShedderStat{name: name}
	go s.run()
	return s
}

// IncrTotal 增加请求总数。
func (s *ShedderStat) IncrTotal() {
	atomic.AddInt64(&s.total, 1)
}

// IncrPass 增加请求通过数。
func (s *ShedderStat) IncrPass() {
	atomic.AddInt64(&s.pass, 1)
}

// IncrDrop 增加请求丢弃数。
func (s *ShedderStat) IncrDrop() {
	atomic.AddInt64(&s.drop, 1)
}

func (s *ShedderStat) run() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	s.loop(ticker.C)
}

func (s *ShedderStat) loop(c <-chan time.Time) {
	for range c {
		result := s.reset()

		// 不记录泄流阀统计项。
		if !logEnabled.True() {
			continue
		}

		cpuUsage := stat.CpuUsage()
		if result.Drop == 0 {
			logx.Statf("(%s) 负载统计 [1m], CPU: %d, 总请求: %d, 通过数: %d, 丢弃数: %d",
				s.name, cpuUsage, result.Total, result.Pass, result.Drop)
		} else {
			logx.Statf("(%s) 降载统计 [1m], CPU: %d, 总请求: %d, 通过数: %d, 丢弃数: %d",
				s.name, cpuUsage, result.Total, result.Pass, result.Drop)
		}
	}
}

// 返回当前统计值并重置为0。
func (s *ShedderStat) reset() snapshot {
	return snapshot{
		Total: atomic.SwapInt64(&s.total, 0),
		Pass:  atomic.SwapInt64(&s.pass, 0),
		Drop:  atomic.SwapInt64(&s.drop, 0),
	}
}
