package clientinterceptors

import (
	"context"

	"git.zc0901.com/go/god/lib/retry"

	"google.golang.org/grpc"
)

// RetryInterceptor 重试拦截器。
func RetryInterceptor(enable bool) grpc.UnaryClientInterceptor {
	if !enable {
		return func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}

	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return retry.Do(ctx, func(ctx context.Context, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}, opts...)
	}
}
