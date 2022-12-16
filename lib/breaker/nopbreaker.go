package breaker

const nopBreakerName = "nopBreaker"

type nopBreaker struct{}

func newNopBreaker() Breaker {
	return nopBreaker{}
}

func (n nopBreaker) Name() string {
	return nopBreakerName
}

func (n nopBreaker) Allow() (Promise, error) {
	return nopPromise{}, nil
}

func (n nopBreaker) Do(req func() error) error {
	return req()
}

func (n nopBreaker) DoWithAcceptable(req func() error, _ Acceptable) error {
	return req()
}

func (n nopBreaker) DoWithFallback(req func() error, _ func(err error) error) error {
	return req()
}

func (n nopBreaker) DoWithFallbackAcceptable(req func() error, _ func(err error) error, _ Acceptable) error {
	return req()
}

type nopPromise struct{}

func (p nopPromise) Accept() {

}

func (p nopPromise) Reject(string) {

}
