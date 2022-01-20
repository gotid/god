package serverinterceptors

import (
	"context"

	"github.com/gotid/god/lib/breaker"
	"github.com/gotid/god/rpc/internal/codes"
	"google.golang.org/grpc"
)

// UnaryBreakerInterceptor 是一个充当断路器的一元拦截器。
func UnaryBreakerInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	breakerName := info.FullMethod
	err = breaker.DoWithAcceptable(breakerName, func() error {
		var err error
		resp, err = handler(ctx, req)
		return err
	}, codes.Acceptable)

	return resp, err
}

// StreamBreakerInterceptor 是一个充当断路器的流式拦截器。
func StreamBreakerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) (err error) {
	breakerName := info.FullMethod
	return breaker.DoWithAcceptable(breakerName, func() error {
		return handler(srv, stream)
	}, codes.Acceptable)
}
