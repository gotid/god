package load

type (
	nopShedder struct{}
	nopPromise struct{}
)

func (p nopPromise) Fail() {}

func (p nopPromise) Pass() {}

func (s nopShedder) Allow() (Promise, error) {
	return nopPromise{}, nil
}

func newNopShedder() Shedder {
	return nopShedder{}
}
