package fx

import "github.com/gotid/god/lib/errorx"

const defaultRetryTimes = 3

type (
	// RetryOption 自定义 DoWithRetry 选项。
	RetryOption func(*retryOptions)

	retryOptions struct {
		times int
	}
)

// DoWithRetry 运行 fn，如有错误则重试。
// 默认重试3次，可自定义重试次数。
func DoWithRetry(fn func() error, opts ...RetryOption) error {
	options := newRetryOptions()
	for _, opt := range opts {
		opt(options)
	}

	var batchErr errorx.BatchError
	for i := 0; i < options.times; i++ {
		if err := fn(); err != nil {
			batchErr.Add(err)
		} else {
			return nil
		}
	}

	return batchErr.Err()
}

// WithRetry 自定义 DoWithRetry 的重试次数。
func WithRetry(times int) RetryOption {
	return func(options *retryOptions) {
		options.times = times
	}
}

func newRetryOptions() *retryOptions {
	return &retryOptions{
		times: defaultRetryTimes,
	}
}
