package metric

import (
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
)

type (
	// HistogramVecOpts 自定义 HistogramVec 的方法。
	HistogramVecOpts struct {
		Namespace string
		Subsystem string
		Name      string
		Help      string
		Labels    []string
		Buckets   []float64
	}

	// HistogramVec 接口代表一个直方图向量。
	HistogramVec interface {
		// Observe 添加观察值 v 到标签 labels。
		Observe(v int64, labels ...string)
		close() bool
	}

	promHistogramVec struct {
		histogram *prom.HistogramVec
	}
)

// NewHistogramVec 返回一个直方图向量 HistogramVec。
func NewHistogramVec(opts *HistogramVecOpts) HistogramVec {
	if opts == nil {
		return nil
	}

	vec := prom.NewHistogramVec(prom.HistogramOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
		Buckets:   opts.Buckets,
	}, opts.Labels)
	prom.MustRegister(vec)
	hv := &promHistogramVec{
		histogram: vec,
	}
	proc.AddShutdownListener(func() {
		hv.close()
	})

	return hv
}

func (hv *promHistogramVec) Observe(v int64, labels ...string) {
	if !prometheus.Enabled() {
		return
	}

	hv.histogram.WithLabelValues(labels...).Observe(float64(v))
}
func (hv *promHistogramVec) close() bool {
	return prom.Unregister(hv.histogram)
}
