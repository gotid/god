package retry

import (
	"context"
	"strconv"
	"time"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/status"

	"git.zc0901.com/go/god/lib/logx"

	"google.golang.org/grpc"

	"git.zc0901.com/go/god/lib/retry/backoff"
	"google.golang.org/grpc/codes"
)

const AttemptMetadataKey = "x-retry-attempt"

var (
	// DefaultRetryCodes 默认重试代码
	DefaultRetryCodes   = []codes.Code{codes.ResourceExhausted, codes.Unavailable}
	defaultRetryOptions = &options{
		max:                0,
		perCallTimeout:     0,
		includeRetryHeader: true,
		codes:              DefaultRetryCodes,
		backoffFunc:        backoff.LinearWithJitter(50*time.Millisecond, 0.10),
	}
)

type (
	// 重试配置项。
	options struct {
		max                int
		perCallTimeout     time.Duration
		includeRetryHeader bool
		codes              []codes.Code
		backoffFunc        backoff.Func
	}

	// CallOption 是一个 grpc.CallOption，用于 grpc 连接重试。
	CallOption struct {
		grpc.EmptyCallOption // 确保实现私有 after() 和 before() 以避免 panic
		apply                func(opt *options)
	}
)

func Do(ctx context.Context, call func(ctx context.Context, opts ...grpc.CallOption) error, opts ...grpc.CallOption) error {
	logger := logx.WithContext(ctx)
	grpcOpts, retryOpts := filterCallOptions(opts)
	callOpts := reuseOrNewWithCallOptions(defaultRetryOptions, retryOpts)

	if callOpts.max == 0 {
		return call(ctx, opts...)
	}

	var lastErr error
	for attempt := 0; attempt <= callOpts.max; attempt++ {
		if err := waitRetryBackoff(logger, attempt, ctx, callOpts); err != nil {
			return err
		}

		callCtx := perCallContext(ctx, callOpts, attempt)
		lastErr = call(callCtx, grpcOpts...)

		if lastErr == nil {
			return nil
		}

		if attempt == 0 {
			logger.Errorf("grpc 调用失败，错误：%v", lastErr)
		} else {
			logger.Errorf("grpc 重试第 %d 次，错误：%v", attempt, lastErr)
		}
		if isContextError(lastErr) {
			if ctx.Err() != nil {
				logger.Errorf("grpc 重试第 %d 次,父级上下文错误：%v", attempt, ctx.Err())
				return lastErr
			} else if callOpts.perCallTimeout != 0 {
				logger.Errorf("grpc 重试第 %d 次，重试错误", attempt)
				continue
			}
		}
		if !canRetry(lastErr, callOpts) {
			return lastErr
		}
	}

	return lastErr
}

func filterCallOptions(callOptions []grpc.CallOption) (grpcOptions []grpc.CallOption, retryOptions []*CallOption) {
	for _, opt := range callOptions {
		if co, ok := opt.(*CallOption); ok {
			retryOptions = append(retryOptions, co)
		} else {
			grpcOptions = append(grpcOptions, opt)
		}
	}

	return grpcOptions, retryOptions
}

func reuseOrNewWithCallOptions(opt *options, retryCallOptions []*CallOption) *options {
	if len(retryCallOptions) == 0 {
		return opt
	}

	return parseRetryCallOptions(opt, retryCallOptions...)
}

func perCallContext(ctx context.Context, callOpts *options, attempt int) context.Context {
	if attempt > 0 {
		if callOpts.perCallTimeout != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, callOpts.perCallTimeout)
			_ = cancel
		}
		if callOpts.includeRetryHeader {
			cloneMd := extractIncomingAndClone(ctx)
			cloneMd.Set(AttemptMetadataKey, strconv.Itoa(attempt))
			ctx = metadata.NewOutgoingContext(ctx, cloneMd)
		}
	}

	return ctx
}

func extractIncomingAndClone(ctx context.Context) metadata.MD {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return metadata.MD{}
	}

	return md.Copy()
}

func parseRetryCallOptions(opt *options, opts ...*CallOption) *options {
	for _, option := range opts {
		option.apply(opt)
	}

	return opt
}

func canRetry(err error, retryOptions *options) bool {
	errCode := status.Code(err)
	if isContextError(err) {
		return false
	}

	for _, code := range retryOptions.codes {
		if code == errCode {
			return true
		}
	}

	return false
}

func isContextError(err error) bool {
	code := status.Code(err)
	return code == codes.DeadlineExceeded || code == codes.Canceled
}

func waitRetryBackoff(logger logx.Logger, attempt int, ctx context.Context, retryOptions *options) error {
	var waitTime time.Duration = 0
	if attempt > 0 {
		waitTime = retryOptions.backoffFunc(attempt)
	}
	if waitTime > 0 {
		timer := time.NewTimer(waitTime)
		defer timer.Stop()

		logger.Infof("grpc 重试第 %d 次，等待时长: %v", attempt, waitTime)
		select {
		case <-ctx.Done():
			return status.FromContextError(ctx.Err()).Err()
		case <-timer.C:
			err := ctx.Err()
			if err != nil {
				return status.FromContextError(err).Err()
			}
		}
	}

	return nil
}
