package serverinterceptors

import (
	"context"
	"strconv"
	"time"

	"git.zc0901.com/go/god/lib/prometheus"
	"git.zc0901.com/go/god/lib/prometheus/metric"
	"git.zc0901.com/go/god/lib/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const serverNamespace = "rpc_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "RPC服务端请求耗时（毫秒）。",
		Labels:    []string{"method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "RPC服务端请求响应码计数器。",
		Labels:    []string{"method", "code"},
	})
)

// UnaryPrometheusInterceptor 统计rpc服务端请求时长和状态代码
func UnaryPrometheusInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !prometheus.Enabled() {
		return handler(ctx, req)
	}

	startTime := timex.Now()
	resp, err := handler(ctx, req)
	metricServerReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), info.FullMethod)
	metricServerReqCodeTotal.Inc(info.FullMethod, strconv.Itoa(int(status.Code(err))))
	return resp, err
}
