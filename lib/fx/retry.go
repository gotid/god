package fx

import "github.com/gotid/god/lib/errorx"

const defaultRetryTimes = 3

type (
	// RetryOption 是一个自定义 DoWithRetry 的函数。
	RetryOption func(*retryOption)

	retryOption struct {
		times int
	}
)

// DoWithRetries 带有重试次数地执行函数
func DoWithRetries(fn func() error, opts ...RetryOption) error {
	options := newRetryOption()
	for _, opt := range opts {
		opt(options)
	}

	var es errorx.Errors
	for i := 0; i < options.times; i++ {
		if err := fn(); err != nil {
			es.Add(err)
		} else {
			return nil
		}
	}

	return es.Error()
}

// WithRetries 自定义重试次数
func WithRetries(times int) RetryOption {
	return func(option *retryOption) {
		option.times = times
	}
}

func newRetryOption() *retryOption {
	return &retryOption{
		times: defaultRetryTimes,
	}
}
