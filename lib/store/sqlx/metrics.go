package sqlx

import "github.com/gotid/god/lib/metric"

const namespace = "sql_client"

var (
	metricReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "MySQL客户端请求时长(毫秒)。",
		Labels:    []string{"command"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500},
	})
	metricReqErr = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "error_total",
		Help:      "MySQL客户端请求错误次数。",
		Labels:    []string{"command", "error"},
	})
)
