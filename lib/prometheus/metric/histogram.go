package metric

import (
	"github.com/gotid/god/lib/proc"
	prom "github.com/prometheus/client_golang/prometheus"
)

type (
	HistogramVecOpts struct {
		Namespace string
		Subsystem string
		Name      string
		Help      string
		Labels    []string
		Buckets   []float64
	}

	HistogramVec interface {
		Observe(v int64, labels ...string)
		close() bool
	}

	promHistogramVec struct {
		histogram *prom.HistogramVec
	}
)

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
	hv.histogram.WithLabelValues(labels...).Observe(float64(v))
}

func (hv *promHistogramVec) close() bool {
	return prom.Unregister(hv.histogram)
}
