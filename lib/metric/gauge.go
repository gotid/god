package metric

import (
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
)

type (
	// GaugeVecOpts 是 VectorOpts 的别名。
	GaugeVecOpts VectorOpts

	// GaugeVec 接口代表一个计量器向量。
	GaugeVec interface {
		// Set 设置值 v 到标签 labels。
		Set(v float64, labels ...string)
		// Inc 递增 labels 次数。
		Inc(labels ...string)
		// Add 添加值 v 到标签 labels。
		Add(v float64, labels ...string)
		close() bool
	}

	promGaugeVec struct {
		gauge *prom.GaugeVec
	}
)

// NewGaugeVec 返回一个计量器向量 GaugeVec。
func NewGaugeVec(cfg *GaugeVecOpts) GaugeVec {
	if cfg == nil {
		return nil
	}

	vec := prom.NewGaugeVec(
		prom.GaugeOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.Subsystem,
			Name:      cfg.Name,
			Help:      cfg.Help,
		}, cfg.Labels)
	prom.MustRegister(vec)
	gv := &promGaugeVec{
		gauge: vec,
	}
	proc.AddShutdownListener(func() {
		gv.close()
	})

	return gv
}

func (gv *promGaugeVec) Inc(labels ...string) {
	if !prometheus.Enabled() {
		return
	}

	gv.gauge.WithLabelValues(labels...).Inc()
}

func (gv *promGaugeVec) Add(v float64, labels ...string) {
	if !prometheus.Enabled() {
		return
	}

	gv.gauge.WithLabelValues(labels...).Add(v)
}

func (gv *promGaugeVec) Set(v float64, labels ...string) {
	if !prometheus.Enabled() {
		return
	}

	gv.gauge.WithLabelValues(labels...).Set(v)
}

func (gv *promGaugeVec) close() bool {
	return prom.Unregister(gv.gauge)
}
