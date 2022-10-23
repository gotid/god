package clientinterceptors

import (
	"context"
	"github.com/gotid/god/lib/metric"
	"github.com/gotid/god/lib/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

const clientNamespace = "rpc_client"

var (
	metricClientReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "RPC客户端请求时长(ms)。",
		Labels:    []string{"method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricClientReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "RPC客户端请求错误次数。",
		Labels:    []string{"method", "code"},
	})
)

// PrometheusInterceptor 报告统计信息给普罗米修斯服务器的客户端拦截器。
func PrometheusInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	startTime := timex.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	metricClientReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), method)
	metricClientReqCodeTotal.Inc(method, strconv.Itoa(int(status.Code(err))))
	return err
}
