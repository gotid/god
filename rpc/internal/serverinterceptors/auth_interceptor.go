package serverinterceptors

import (
	"context"

	"github.com/gotid/god/rpc/internal/auth"
	"google.golang.org/grpc"
)

func UnaryAuthorizeInterceptor(authenticator *auth.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if err := authenticator.Authenticate(ctx); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func StreamAuthorizeInterceptor(authenticator *auth.Authenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authenticator.Authenticate(stream.Context()); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}
