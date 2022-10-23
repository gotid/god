package serverinterceptors

import (
	"context"
	"github.com/gotid/god/lib/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime/debug"
)

// UnaryCrashInterceptor 用于一元请求的异常捕获拦截器。
func UnaryCrashInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer handleCrash(func(r interface{}) {
		err = toPanicError(r)
	})

	return handler(ctx, req)
}

// StreamCrashInterceptor 捕获 stream 请求和 recover() 中的 panics。
func StreamCrashInterceptor(svr interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer handleCrash(func(r interface{}) {
		err = toPanicError(r)
	})

	return handler(svr, stream)
}

func toPanicError(r interface{}) error {
	logx.Errorf("%+v\n\n%s", r, debug.Stack())
	return status.Errorf(codes.Internal, "panic: %v", r)
}

func handleCrash(handler func(interface{})) {
	if r := recover(); r != nil {
		handler(r)
	}
}
