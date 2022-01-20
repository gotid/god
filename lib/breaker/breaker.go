package breaker

import (
	"errors"

	"github.com/gotid/god/lib/stringx"
)

// ErrServiceUnavailable 当 Breaker 打开时返回的错误信息
var ErrServiceUnavailable = errors.New("断路器打开")

type (
	Acceptable func(reqError error) bool   // 判断错误是否可接受的函数
	Request    func() error                // 待执行的请求
	Fallback   func(acceptErr error) error // 备用函数
	Option     func(b *breaker)            // 自定义断路器的方法

	// Breaker 表示一个断路器。
	Breaker interface {
		// Name 是 netflixBreaker 断路器的名称
		Name() string

		// Allow 检查请求是否被允许。
		// 若允许，则返回 Promise，调用者需在成功时调用 promise.Resolve()，失败时调用 promise.Reject()。
		// 若不允许，则返回 ErrServiceUnavailable。
		Allow() (Promise, error)

		// Do 若断路器允许则执行请求，否则返回错误。
		Do(req Request) error

		// DoWithFallback 若断路器允许则执行请求，反之则调用备用函数，再之返回错误。
		DoWithFallback(req Request, fallback Fallback) error

		// DoWithAcceptable 若断路器允许则执行请求，反之则返回错误，并判断错误可否标记为成功请求。
		DoWithAcceptable(req Request, acceptable Acceptable) error

		// DoWithFallbackAcceptable 若断路器允许则执行请求，反之则调用备用函数，再之返回错误并判断错误可否标记为成功请求。
		DoWithFallbackAcceptable(req Request, fallback Fallback, acceptable Acceptable) error
	}

	// Promise 接口定义 Breaker.Allow 返回的回调方法。
	Promise interface {
		Accept()              // 告知 Breaker 调用成功。
		Reject(reason string) // 告知 Breaker 调用失败。
	}

	throttle interface {
		allow() (Promise, error)
		doReq(req Request, fallback Fallback, acceptable Acceptable) error
	}

	breaker struct {
		name string
		throttle
	}
)

// NewBreaker 返回一个 Breaker 断路器对象。
// opts 用于自定义 Breaker。
func NewBreaker(opts ...Option) Breaker {
	var b breaker
	for _, opt := range opts {
		opt(&b)
	}
	if len(b.name) == 0 {
		b.name = stringx.Rand()
	}
	b.throttle = newLoggedThrottle(b.name, newGoogleBreaker())
	return &b
}

// WithName 返回设置 Breaker 名称的函数。
func WithName(name string) Option {
	return func(b *breaker) {
		b.name = name
	}
}

func (b *breaker) Name() string {
	return b.name
}

func (b *breaker) Allow() (Promise, error) {
	return b.throttle.allow()
}

func (b *breaker) Do(req Request) error {
	return b.throttle.doReq(req, nil, defaultAcceptable)
}

func (b *breaker) DoWithFallback(req Request, fallback Fallback) error {
	return b.throttle.doReq(req, fallback, defaultAcceptable)
}

func (b *breaker) DoWithAcceptable(req Request, acceptable Acceptable) error {
	return b.throttle.doReq(req, nil, acceptable)
}

func (b *breaker) DoWithFallbackAcceptable(req Request, fallback Fallback, acceptable Acceptable) error {
	return b.throttle.doReq(req, fallback, acceptable)
}

func defaultAcceptable(err error) bool {
	return err == nil
}
