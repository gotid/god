package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gotid/god/api/internal/security"
	"github.com/gotid/god/lib/prometheus"
	"github.com/gotid/god/lib/prometheus/metric"
	"github.com/gotid/god/lib/timex"
)

const serverNamespace = "http_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "HTTP 服务器请求耗时（毫秒）。",
		Labels:    []string{"path"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "HTTP 服务器请求错误计数器。",
		Labels:    []string{"path", "code"},
	})
)

// PrometheusHandler 返回一个上报状态给普罗米修斯的中间件。
func PrometheusHandler(path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !prometheus.Enabled() {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := timex.Now()
			cw := &security.WithCodeResponseWriter{Writer: w}
			defer func() {
				metricServerReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), path)
				metricServerReqCodeTotal.Inc(path, strconv.Itoa(cw.Code))
			}()

			next.ServeHTTP(cw, r)
		})
	}
}
