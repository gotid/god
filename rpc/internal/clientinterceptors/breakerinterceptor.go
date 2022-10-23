package clientinterceptors

import (
	"context"
	"github.com/gotid/god/lib/breaker"
	"github.com/gotid/god/rpc/internal/codes"
	"google.golang.org/grpc"
	"path"
)

// BreakerInterceptor 是一个作为断路器的客户端拦截器。
func BreakerInterceptor(ctx context.Context, method string, req, reply interface{},
	conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	breakerName := path.Join(conn.Target(), method)
	return breaker.DoWithAcceptable(breakerName, func() error {
		return invoker(ctx, method, req, reply, conn, opts...)
	}, codes.Acceptable)
}
