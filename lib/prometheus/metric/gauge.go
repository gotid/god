package metric

import (
	"github.com/gotid/god/lib/proc"
	prom "github.com/prometheus/client_golang/prometheus"
)

type (
	GaugeVecOpts VectorOpts

	GaugeVec interface {
		Set(v float64, labels ...string)
		Inc(labels ...string)
		Add(v float64, labels ...string)
		close() bool
	}

	promGaugeVec struct {
		gauge *prom.GaugeVec
	}
)

func (gv *promGaugeVec) Set(v float64, labels ...string) {
	gv.gauge.WithLabelValues(labels...).Set(v)
}

func (gv *promGaugeVec) Inc(labels ...string) {
	gv.gauge.WithLabelValues(labels...).Inc()
}

func (gv *promGaugeVec) Add(v float64, labels ...string) {
	gv.gauge.WithLabelValues(labels...).Add(v)
}

func (gv *promGaugeVec) close() bool {
	return prom.Unregister(gv.gauge)
}

func NewGaugeVec(opts *GaugeVecOpts) GaugeVec {
	if opts == nil {
		return nil
	}

	vec := prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
	}, opts.Labels)
	prom.MustRegister(vec)
	gv := &promGaugeVec{gauge: vec}
	proc.AddShutdownListener(func() {
		gv.close()
	})

	return gv
}
