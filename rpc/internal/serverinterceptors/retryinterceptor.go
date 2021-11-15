package serverinterceptors

import (
	"context"
	"strconv"

	"git.zc0901.com/go/god/lib/logx"

	"git.zc0901.com/go/god/lib/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RetryInterceptor 连接重连拦截器。
func RetryInterceptor(maxRetries int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		var md metadata.MD
		reqMd, ok := metadata.FromIncomingContext(ctx)
		if ok {
			md = reqMd.Copy()
			attemptMd := md.Get(retry.AttemptMetadataKey)
			if len(attemptMd) != 0 && attemptMd[0] != "" {
				if attempt, err := strconv.Atoi(attemptMd[0]); err == nil {
					if attempt > maxRetries {
						logx.WithContext(ctx).Errorf("超过重试次数：%d，最大允许重试次数：%d", attempt, maxRetries)
						return nil, status.Error(codes.FailedPrecondition, "超过重试次数")
					}
				}
			}
		}

		return handler(ctx, req)
	}
}
