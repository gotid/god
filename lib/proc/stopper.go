package proc

var nopStopper nilStopper

type (
	// Stopper 接口包装 Stop 方法。
	Stopper interface {
		Stop()
	}

	nilStopper struct{}
)

func (n nilStopper) Stop() {}
