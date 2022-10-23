package serverinterceptors

import (
	"context"
	"github.com/gotid/god/rpc/internal/auth"
	"google.golang.org/grpc"
)

// UnaryAuthorizeInterceptor 用于一元请求的鉴权拦截器。
func UnaryAuthorizeInterceptor(authenticator *auth.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := authenticator.Authenticate(ctx); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// StreamAuthorizeInterceptor 用于流式请求的鉴权拦截器。
func StreamAuthorizeInterceptor(authenticator *auth.Authenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authenticator.Authenticate(stream.Context()); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}
