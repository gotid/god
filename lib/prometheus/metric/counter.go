package metric

import (
	"github.com/gotid/god/lib/proc"
	prom "github.com/prometheus/client_golang/prometheus"
)

type (
	CounterVecOpts VectorOpts

	CounterVec interface {
		Inc(labels ...string)
		Add(v float64, labels ...string)
		close() bool
	}

	promCounterVec struct {
		counter *prom.CounterVec
	}
)

func (cv *promCounterVec) Inc(labels ...string) {
	cv.counter.WithLabelValues(labels...).Inc()
}

func (cv *promCounterVec) Add(v float64, labels ...string) {
	cv.counter.WithLabelValues(labels...).Add(v)
}

func (cv *promCounterVec) close() bool {
	return prom.Unregister(cv.counter)
}

func NewCounterVec(opts *CounterVecOpts) CounterVec {
	if opts == nil {
		return nil
	}

	vec := prom.NewCounterVec(prom.CounterOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
	}, opts.Labels)
	prom.MustRegister(vec)
	cv := &promCounterVec{counter: vec}
	proc.AddShutdownListener(func() {
		cv.close()
	})

	return cv
}
