package handler

import (
	"github.com/gotid/god/api/internal/response"
	"github.com/gotid/god/lib/metric"
	"github.com/gotid/god/lib/timex"
	"net/http"
	"strconv"
	"time"
)

const serverNamespace = "http_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "HTTP服务器请求时长（毫秒）。",
		Labels:    []string{"path"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "HTTP服务器请求错误次数。",
		Labels:    []string{"path", "code"},
	})
)

// PrometheusHandler 返回一个上报请求时长和错误次数给普罗米修斯的数据统计中间件。
func PrometheusHandler(path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := timex.Now()
			cw := &response.WithCodeResponseWriter{Writer: w}
			defer func() {
				metricServerReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), path)
				metricServerReqCodeTotal.Inc(path, strconv.Itoa(cw.Code))
			}()

			next.ServeHTTP(cw, r)
		})
	}
}
