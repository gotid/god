package clientinterceptors

import (
	"context"
	"strconv"
	"time"

	"github.com/gotid/god/lib/prometheus"
	"github.com/gotid/god/lib/prometheus/metric"
	"github.com/gotid/god/lib/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const clientNamespace = "rpc_client"

var (
	metricClientReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "RPC客户端请求耗时（毫秒）。",
		Labels:    []string{"method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricClientReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "RPC客户端请求响应码计数器。",
		Labels:    []string{"method", "code"},
	})
)

func PrometheusInterceptor(ctx context.Context, method string, req, reply interface{},
	conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if !prometheus.Enabled() {
		return invoker(ctx, method, req, reply, conn, opts...)
	}

	startTime := timex.Now()
	err := invoker(ctx, method, req, reply, conn, opts...)
	metricClientReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), method)
	metricClientReqCodeTotal.Inc(method, strconv.Itoa(int(status.Code(err))))
	return err
}
