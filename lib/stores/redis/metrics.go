package redis

import "github.com/gotid/god/lib/metric"

const namespace = "redis_client"

var (
	// redis客户端请求时长直方图指标
	// 某个命令在某个桶中
	metricReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "Redis客户端请求时长（毫秒）。",
		Labels:    []string{"command"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500},
	})

	// redis客户端请求错误次数指标
	// 某个命令在某种错误的次数
	metricReqErr = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "error_total",
		Help:      "Redis客户端请求错误次数。",
		Labels:    []string{"command", "error"},
	})
)
