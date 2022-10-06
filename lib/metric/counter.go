package metric

import (
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
)

type (
	// CounterVecOpts 是 VectorOpts 的别名。
	CounterVecOpts VectorOpts

	// CounterVec 接口代表一个计数器向量。
	CounterVec interface {
		// Inc 递增 labels 次数。
		Inc(labels ...string)
		// Add 添加值 v 到标签 labels。
		Add(v float64, labels ...string)
		close() bool
	}

	promCounterVec struct {
		counter *prom.CounterVec
	}
)

// NewCounterVec 返回一个计数器向量 CounterVec。
func NewCounterVec(opt *CounterVecOpts) CounterVec {
	if opt == nil {
		return nil
	}

	vec := prom.NewCounterVec(prom.CounterOpts{
		Namespace: opt.Namespace,
		Subsystem: opt.Subsystem,
		Name:      opt.Name,
		Help:      opt.Name,
	}, opt.Labels)
	prom.MustRegister(vec)
	cv := &promCounterVec{
		counter: vec,
	}
	proc.AddShutdownListener(func() {
		cv.close()
	})

	return cv
}

func (cv *promCounterVec) Inc(labels ...string) {
	if !prometheus.Enabled() {
		return
	}

	cv.counter.WithLabelValues(labels...).Inc()
}

func (cv *promCounterVec) Add(v float64, labels ...string) {
	if !prometheus.Enabled() {
		return
	}

	cv.counter.WithLabelValues(labels...).Add(v)
}

func (cv *promCounterVec) close() bool {
	return prom.Unregister(cv.counter)
}
