package clientinterceptors

import (
	"context"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mapping"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
	"google.golang.org/grpc"
	"path"
	"time"
)

const defaultSlowThreshold = 500 * time.Millisecond

var slowThreshold = syncx.ForAtomicDuration(defaultSlowThreshold)

// DurationInterceptor 用于时长记录的客户端拦截器。
func DurationInterceptor(ctx context.Context, method string, req, reply interface{},
	conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	serverName := path.Join(conn.Target(), method)
	start := timex.Now()
	err := invoker(ctx, method, req, reply, conn, opts...)
	if err != nil {
		logx.WithContext(ctx).WithDuration(timex.Since(start)).Errorf("失败 - %s - %v - %s",
			serverName, mapping.Repr(req), err.Error())
	} else {
		elapsed := timex.Since(start)
		if elapsed > slowThreshold.Load() {
			logx.WithContext(ctx).WithDuration(elapsed).Slowf("[RPC] 成功 - 慢调用 - %s - %v - %s",
				serverName, mapping.Repr(req), reply)
		}
	}

	return err
}

// SetSlowThreshold 设置慢调用时长阈值。
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold.Set(threshold)
}
