package breaker

import "sync"

var (
	lock     sync.RWMutex
	breakers = make(map[string]Breaker)
)

// Do 使用给定的名称调用 Breaker.Do。
func Do(name string, req func() error) error {
	return do(name, func(b Breaker) error {
		return b.Do(req)
	})
}

// DoWithFallback 使用给定的名称调用 Breaker.DoWithFallback。
func DoWithFallback(name string, req func() error, fallback func(err error) error) error {
	return do(name, func(b Breaker) error {
		return b.DoWithFallback(req, fallback)
	})
}

// DoWithFallbackAcceptable 使用给定的名称调用 Breaker.DoWithFallbackAcceptable。
func DoWithFallbackAcceptable(name string, req func() error, fallback func(err error) error, acceptable Acceptable) error {
	return do(name, func(b Breaker) error {
		return b.DoWithFallbackAcceptable(req, fallback, acceptable)
	})
}

// GetBreaker 返回给定名称的断路器。
func GetBreaker(name string) Breaker {
	lock.RLock()
	b, ok := breakers[name]
	lock.RUnlock()
	if ok {
		return b
	}

	lock.Lock()
	b, ok = breakers[name]
	if !ok {
		b = New(WithName(name))
		breakers[name] = b
	}
	lock.Unlock()

	return b
}

// NoBreakerFor 禁用给定名称的断路器。
func NoBreakerFor(name string) {
	lock.Lock()
	defer lock.Unlock()
	breakers[name] = newNopBreaker()
}

func do(name string, execute func(b Breaker) error) error {
	return execute(GetBreaker(name))
}
