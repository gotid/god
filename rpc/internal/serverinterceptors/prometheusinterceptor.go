package serverinterceptors

import (
	"context"
	"github.com/gotid/god/lib/metric"
	"github.com/gotid/god/lib/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

const serverNamespace = "rpc_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "RPC服务器请求时长(ms)。",
		Labels:    []string{"method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "RPC客户端请求错误次数。",
		Labels:    []string{"method", "code"},
	})
)

// UnaryPrometheusInterceptor 报告统计信息给普罗米修斯服务器的服务端拦截器。
func UnaryPrometheusInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := timex.Now()
	resp, err := handler(ctx, req)
	metricServerReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), info.FullMethod)
	metricServerReqCodeTotal.Inc(info.FullMethod, strconv.Itoa(int(status.Code(err))))
	return resp, err
}
