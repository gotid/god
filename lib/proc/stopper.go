package proc

var nopStopper nilStopper

type (
	// Stopper 表示一个带有停止方法的接口。
	Stopper interface {
		Stop()
	}

	nilStopper struct{}
)

func (s nilStopper) Stop() {}
