package load

// nopShedder 无操作的负载泄流阀。
type nopShedder struct{}

func newNopShedder() nopShedder {
	return nopShedder{}
}

func (s nopShedder) Allow() (Promise, error) {
	return nopPromise{}, nil
}

type nopPromise struct{}

func (p nopPromise) Pass() {}

func (p nopPromise) Fail() {}
