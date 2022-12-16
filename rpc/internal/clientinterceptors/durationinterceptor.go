package clientinterceptors

import (
	"context"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
	"google.golang.org/grpc"
	"path"
	"sync"
	"time"
)

const defaultSlowThreshold = 500 * time.Millisecond

var (
	notLoggingContentMethods sync.Map // 存储无需记录处理时长的方法
	slowThreshold            = syncx.ForAtomicDuration(defaultSlowThreshold)
)

// DurationInterceptor 用于记录处理时长的客户端拦截器。
func DurationInterceptor(ctx context.Context, method string, req, reply interface{},
	conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	serverName := path.Join(conn.Target(), method)
	start := timex.Now()
	err := invoker(ctx, method, req, reply, conn, opts...)
	if err != nil {
		logger := logx.WithContext(ctx).WithDuration(timex.Since(start))
		_, ok := notLoggingContentMethods.Load(method)
		if ok {
			logger.Errorf("失败 - %s - %s", serverName, err.Error())
		} else {
			logger.Errorf("失败 - %s - %v - %s", serverName, req, err.Error())
		}
	} else {
		elapsed := timex.Since(start)
		if elapsed > slowThreshold.Load() {
			logger := logx.WithContext(ctx).WithDuration(elapsed)
			_, ok := notLoggingContentMethods.Load(method)
			if ok {
				logger.Slowf("[RPC] 成功 - 慢调用 - %s", serverName)
			} else {
				logger.Slowf("[RPC] 成功 - 慢调用 - %s - %v - %s", serverName, req, reply)
			}
		}
	}

	return err
}

// DontLogContentMethod 不再记录给定方法的请求/响应详情。
func DontLogContentMethod(method string) {
	notLoggingContentMethods.Store(method, lang.Placeholder)
}

// SetSlowThreshold 设置慢调用时长阈值。
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold.Set(threshold)
}
