package breaker

import (
	"fmt"

	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/stat"
)

type (
	internalPromise interface {
		Accept()
		Reject()
	}

	// 内部节流阀接口，如 google_throttle 实现
	internalThrottle interface {
		allow() (internalPromise, error)
		doReq(req Request, fallback Fallback, acceptable Acceptable) error
	}

	promiseWithReason struct {
		promise internalPromise
		errWin  *errorWindow
	}

	loggedThrottle struct {
		name string
		internalThrottle
		errWin *errorWindow
	}
)

func newLoggedThrottle(name string, t internalThrottle) loggedThrottle {
	return loggedThrottle{
		name:             name,
		internalThrottle: t,
		errWin:           new(errorWindow),
	}
}

func (t loggedThrottle) allow() (Promise, error) {
	promise, err := t.internalThrottle.allow()
	return promiseWithReason{
		promise: promise,
		errWin:  t.errWin,
	}, t.logError(err)
}

func (t loggedThrottle) doReq(req Request, fallback Fallback, acceptable Acceptable) error {
	return t.logError(t.internalThrottle.doReq(req, fallback, func(err error) bool {
		accept := acceptable(err)
		if !accept {
			t.errWin.add(err.Error())
		}
		return accept
	}))
}

func (t loggedThrottle) logError(err error) error {
	if err == ErrServiceUnavailable {
		stat.Report(fmt.Sprintf(
			"进程(%s/%d), 调用者名称: %s, 断路器已打开，请求被丢弃\n最后错误：\n%s",
			proc.ProcessName(), proc.Pid(), t.name, t.errWin))
	}
	return err
}

func (p promiseWithReason) Accept() {
	p.promise.Accept()
}

func (p promiseWithReason) Reject(reason string) {
	p.errWin.add(reason)
	p.promise.Reject()
}
