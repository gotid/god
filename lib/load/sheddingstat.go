package load

import (
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
	"sync/atomic"
	"time"
)

type (
	// SheddingStat 用于存储自动降载的统计信息。
	SheddingStat struct {
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

// NewSheddingStat 返回一个泄流器统计 SheddingStat。
func NewSheddingStat(name string) *SheddingStat {
	st := &SheddingStat{name: name}
	go st.run()

	return st
}

// IncrTotal 增加总请求数。
func (s *SheddingStat) IncrTotal() {
	atomic.AddInt64(&s.total, 1)
}

// IncrPass 增加通过数。
func (s *SheddingStat) IncrPass() {
	atomic.AddInt64(&s.pass, 1)
}

// IncrDrop 增加丢弃数。
func (s *SheddingStat) IncrDrop() {
	atomic.AddInt64(&s.drop, 1)
}

func (s *SheddingStat) run() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	s.loop(ticker.C)
}

func (s *SheddingStat) loop(c <-chan time.Time) {
	for range c {
		st := s.reset()

		if !logEnabled.True() {
			continue
		}

		cpuUsage := stat.CpuUsage()
		if st.Drop == 0 {
			logx.Statf("(%s) 负载统计 [1m]，CPU：%d，请求：%d，通过：%d，丢弃：%d",
				s.name, cpuUsage, st.Total, st.Pass, st.Drop)
		} else {
			logx.Statf("(%s) 负载统计_丢弃 [1m]，CPU：%d，请求：%d，通过：%d，丢弃：%d",
				s.name, cpuUsage, st.Total, st.Pass, st.Drop)
		}
	}
}

func (s *SheddingStat) reset() snapshot {
	return snapshot{
		Total: atomic.SwapInt64(&s.total, 0),
		Pass:  atomic.SwapInt64(&s.pass, 0),
		Drop:  atomic.SwapInt64(&s.drop, 0),
	}
}
