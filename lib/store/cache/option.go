package cache

import "time"

const (
	defaultExpire         = 7 * 24 * time.Hour
	defaultNotFoundExpire = time.Minute
)

type (
	// Options 用于存储缓存选项。
	Options struct {
		Expire         time.Duration
		NotFoundExpire time.Duration
	}

	// Option 自定义缓存选项 Options。
	Option func(options *Options)
)

func newOptions(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}

	if o.Expire <= 0 {
		o.Expire = defaultExpire
	}
	if o.NotFoundExpire <= 0 {
		o.NotFoundExpire = defaultNotFoundExpire
	}

	return o
}

// WithExpire 返回一个自定义缓存选项有效期的函数。
func WithExpire(expire time.Duration) Option {
	return func(o *Options) {
		o.Expire = expire
	}
}

// WithNotFoundExpire 返回一个自定义缓存选项未找到记录的有效期的函数。
func WithNotFoundExpire(expire time.Duration) Option {
	return func(o *Options) {
		o.NotFoundExpire = expire
	}
}
