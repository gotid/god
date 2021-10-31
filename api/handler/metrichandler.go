package handler

import (
	"net/http"

	"git.zc0901.com/go/god/lib/stat"
	"git.zc0901.com/go/god/lib/timex"
)

// MetricHandler 返回一个统计耗时指标的中间件。
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
