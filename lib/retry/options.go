package retry

import (
	"time"

	"git.zc0901.com/go/god/lib/retry/backoff"
	"google.golang.org/grpc/codes"
)

// WithDisable 禁用调用或拦截器的重试行为。
func WithDisable() *CallOption {
	return WithMax(0)
}

// WithMax 设置调用或拦截器的最大重试次数。
func WithMax(maxRetries int) *CallOption {
	return &CallOption{apply: func(opt *options) {
		opt.max = maxRetries
	}}
}

// WithBackoff 设置用于控制重试时间的 backoff.Func。
func WithBackoff(backoffFunc backoff.Func) *CallOption {
	return &CallOption{apply: func(opt *options) {
		opt.backoffFunc = backoffFunc
	}}
}

// WithCodes 设置可重试的 codes.Code。
func WithCodes(retryCodes ...codes.Code) *CallOption {
	return &CallOption{apply: func(opt *options) {
		opt.codes = retryCodes
	}}
}

// WithPerRetryTimeout 设置每次重试的超时时长。
func WithPerRetryTimeout(timeout time.Duration) *CallOption {
	return &CallOption{apply: func(opt *options) {
		opt.perCallTimeout = timeout
	}}
}
