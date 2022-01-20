package serverinterceptors

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gotid/god/lib/syncx"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// 慢日志阈值
const defaultSlowThreshold = 500 * time.Millisecond

var slowThreshold = syncx.ForAtomicDuration(defaultSlowThreshold)

// SetSlowThreshold 设置服务端慢调用时长阈值。
func SetSlowThreshold(duration time.Duration) {
	slowThreshold.Set(duration)
}

// UnaryStatInterceptor 一元统计拦截器（统计请求地址-方法-入参，时长等信息）
func UnaryStatInterceptor(metrics *stat.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer handleCrash(func(r interface{}) {
			err = toPanicError(r)
		})

		startTime := timex.Now()
		defer func() {
			duration := timex.Since(startTime)
			metrics.Add(stat.Task{
				Duration: duration,
			}) // 通过拦截器添加监控指标
			logDuration(ctx, info.FullMethod, req, duration)
		}()

		return handler(ctx, req)
	}
}

func logDuration(ctx context.Context, method string, req interface{}, duration time.Duration) {
	var addr string
	client, ok := peer.FromContext(ctx)
	if ok {
		addr = client.Addr.String()
	}
	content, err := json.Marshal(req)
	if err != nil {
		logx.WithContext(ctx).Errorf("%s - %s", addr, err.Error())
	} else if duration > defaultSlowThreshold {
		logx.WithContext(ctx).WithDuration(duration).Slowf("[RPC] 慢调用 - %s - %s - %s",
			addr, method, string(content))
	} else {
		logx.WithContext(ctx).WithDuration(duration).Infof("%s - %s - %s", addr, method, string(content))
	}
}
