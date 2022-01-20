package clientinterceptors

import (
	"context"
	"path"
	"time"

	"git.zc0901.com/go/god/lib/syncx"

	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/timex"
	"google.golang.org/grpc"
)

const defaultSlowThreshold = 500 * time.Millisecond

var slowThreshold = syncx.ForAtomicDuration(defaultSlowThreshold)

// DurationInterceptor rpc调用时长拦截器
func DurationInterceptor(ctx context.Context, method string, req, replay interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	serviceName := path.Join(cc.Target(), method)
	start := timex.Now()
	err := invoker(ctx, method, req, replay, cc, opts...)
	if err != nil {
		logx.WithContext(ctx).WithDuration(timex.Since(start)).Errorf("失败 - %s - %v - %s",
			serviceName, req, err.Error())
	} else {
		elapsed := timex.Since(start)
		if elapsed > slowThreshold.Load() {
			logx.WithContext(ctx).WithDuration(elapsed).Slowf("OK - 慢调用 - %s -%v - %v",
				serviceName, req, replay)
		}
	}

	return err
}

// SetSlowThreshold 设置慢调用时间阈值。
func SetSlowThreshold(duration time.Duration) {
	slowThreshold.Set(duration)
}
