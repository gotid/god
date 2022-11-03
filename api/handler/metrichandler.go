package handler

import (
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/timex"
	"net/http"
)

// MetricHandler 返回一个请求时长的指标统计中间件。
func MetricHandler(metrics *stat.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := timex.Now()
			defer func() {
				metrics.Add(stat.Task{
					Duration: timex.Since(startTime),
				})
			}()

			next.ServeHTTP(w, r)
		})
	}
}
