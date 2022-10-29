package breaker

import (
	"errors"
	"fmt"
	"github.com/gotid/god/lib/mathx"
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/stringx"
	"strings"
	"sync"
	"time"
)

const (
	numHistoryReasons = 5
	timeFormat        = "15:04:05"
)

// ErrServiceUnavailable 代表断路器状态为打开时的错误。
var ErrServiceUnavailable = errors.New("断路器已打开")

type (
	// Acceptable 是一个用于检查错误是否可被接收的函数。
	Acceptable func(err error) bool

	// Promise 接口定义 Breaker.Allow 返回的回调函数。
	Promise interface {
		// Accept 告知断路器调用成功。
		Accept()
		// Reject 告知断路器调用失败。
		Reject(reason string)
	}

	// Breaker 表示一个可自动熔断的断路器。
	Breaker interface {
		// Name 返回断路器的名称。
		Name() string

		// Allow 检查请求是否被允许。
		// 如果允许，将返回一个 Promise，调用者需要在成功时调用 Promise.Accept()，
		// 失败时调用 Promise.Reject()。
		// 如果不允许，将返回 ErrServiceUnavailable。
		Allow() (Promise, error)

		// Do 如果断路器允许，执行给定的请求 req，反之则返回错误。
		Do(req func() error) error

		// DoWithAcceptable 如果断路器允许，执行给定的请求 req，反之则返回错误。
		// 错误是否可接受，由接受度检查函数进行确定。
		DoWithAcceptable(req func() error, acceptable Acceptable) error

		// DoWithFallback 如果断路器允许，执行给定的请求 req，如果拒绝则执行 fallback。
		DoWithFallback(req func() error, fallback func(err error) error) error

		// DoWithFallbackAcceptable 如果断路器允许，执行给定的请求 req，如果拒绝则执行 fallback。
		// 错误是否可接受，由接受度检查函数进行确定。
		DoWithFallbackAcceptable(req func() error, fallback func(err error) error, acceptable Acceptable) error
	}

	// Option 自定义断路器的方法。
	Option func(breaker *circuitBreaker)

	circuitBreaker struct {
		name string
		throttle
	}

	throttle interface {
		allow() (Promise, error)
		doReq(req func() error, fallback func(err error) error, acceptable Acceptable) error
	}

	internalPromise interface {
		Accept()
		Reject()
	}

	internalThrottle interface {
		allow() (internalPromise, error)
		doReq(req func() error, fallback func(err error) error, acceptable Acceptable) error
	}
)

// New 返回一个给定可选项的 Breaker 实例。
func New(opts ...Option) Breaker {
	var b circuitBreaker
	for _, opt := range opts {
		opt(&b)
	}
	if len(b.name) == 0 {
		b.name = stringx.Rand()
	}
	b.throttle = newLoggedThrottle(b.name, newGoogleBreaker())

	return &b
}

func (cb *circuitBreaker) Name() string {
	return cb.name
}

func (cb *circuitBreaker) Allow() (Promise, error) {
	return cb.throttle.allow()
}

func (cb *circuitBreaker) Do(req func() error) error {
	return cb.throttle.doReq(req, nil, defaultAcceptable)
}

func (cb *circuitBreaker) DoWithAcceptable(req func() error, acceptable Acceptable) error {
	return cb.throttle.doReq(req, nil, acceptable)
}

func (cb *circuitBreaker) DoWithFallback(req func() error, fallback func(err error) error) error {
	return cb.throttle.doReq(req, fallback, defaultAcceptable)
}

func (cb *circuitBreaker) DoWithFallbackAcceptable(req func() error, fallback func(err error) error, acceptable Acceptable) error {
	return cb.throttle.doReq(req, fallback, acceptable)
}

// WithName 返回自定义断路器名称的可选项函数。
func WithName(name string) Option {
	return func(breaker *circuitBreaker) {
		breaker.name = name
	}
}
func defaultAcceptable(err error) bool {
	return err == nil
}

type loggedThrottle struct {
	name string
	internalThrottle
	errWin *errorWindow
}

func newLoggedThrottle(name string, t internalThrottle) loggedThrottle {
	return loggedThrottle{
		name:             name,
		internalThrottle: t,
		errWin:           new(errorWindow),
	}
}

func (lt loggedThrottle) allow() (Promise, error) {
	promise, err := lt.internalThrottle.allow()
	return promiseWithReason{
		promise: promise,
		errWin:  lt.errWin,
	}, lt.logError(err)
}

func (lt loggedThrottle) doReq(req func() error, fallback func(err error) error, acceptable Acceptable) error {
	return lt.logError(lt.internalThrottle.doReq(req, fallback, func(err error) bool {
		accept := acceptable(err)
		if !accept && err != nil {
			lt.errWin.add(err.Error())
		}
		return accept
	}))
}

func (lt loggedThrottle) logError(err error) error {
	if err == ErrServiceUnavailable {
		// 如果错误是断路器打开，则一定会有错误窗口
		stat.Report(fmt.Sprintf(
			"proc(%s/%d), caller: %s, 断路器已打开且请求已丢弃\n最新错误：\n%s",
			proc.ProcessName(), proc.Pid(), lt.name, lt.errWin))
	}

	return err
}

type errorWindow struct {
	reasons [numHistoryReasons]string
	index   int
	count   int
	lock    sync.Mutex
}

func (ew *errorWindow) add(reason string) {
	ew.lock.Lock()
	defer ew.lock.Unlock()

	ew.reasons[ew.index] = fmt.Sprintf("%s %s", time.Now().Format(timeFormat), reason)
	ew.index = (ew.index + 1) % numHistoryReasons
	ew.count = mathx.MinInt(ew.count+1, numHistoryReasons)
}

func (ew *errorWindow) String() string {
	var reasons []string

	ew.lock.Lock()
	defer ew.lock.Unlock()

	// 反向顺序
	for i := ew.index - 1; i >= ew.index-ew.count; i-- {
		reasons = append(reasons, ew.reasons[(i+numHistoryReasons)%numHistoryReasons])
	}

	return strings.Join(reasons, "\n")
}

type promiseWithReason struct {
	promise internalPromise
	errWin  *errorWindow
}

func (p promiseWithReason) Accept() {
	p.promise.Accept()
}

func (p promiseWithReason) Reject(reason string) {
	p.errWin.add(reason)
	p.promise.Reject()
}
