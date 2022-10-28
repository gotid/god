package serverinterceptors

import (
	"context"
	"encoding/json"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"sync"
	"time"
)

const defaultSlowThreshold = 500 * time.Millisecond

var (
	notLoggingContentMethods sync.Map
	slowThreshold            = syncx.ForAtomicDuration(defaultSlowThreshold)
)

// SetSlowThreshold 设置慢阈值。
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold.Set(threshold)
}

// DontLogContentForMethod 禁用给定方法的日志内容。
func DontLogContentForMethod(method string) {
	notLoggingContentMethods.Store(method, lang.Placeholder)
}

// UnaryStatInterceptor 返回给定指标的函数来汇报统计信息。
func UnaryStatInterceptor(metrics *stat.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := timex.Now()
		defer func() {
			duration := timex.Since(startTime)
			metrics.Add(stat.Task{
				Duration: duration,
			})
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

	logger := logx.WithContext(ctx).WithDuration(duration)

	if _, ok = notLoggingContentMethods.Load(method); ok {
		if duration > slowThreshold.Load() {
			logger.Slowf("[RPC] 慢调用 - %s - %s", addr, method)
		}
	} else {
		content, err := json.Marshal(req)
		if err != nil {
			logx.WithContext(ctx).Errorf("%s - %s", addr, err.Error())
		} else if duration > slowThreshold.Load() {
			logger.Slowf("[RPC] 慢调用 - %s - %s - %s", addr, method, string(content))
		} else {
			logger.Infof("%s - %s - %s", addr, method, string(content))
		}
	}
}
