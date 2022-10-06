package timex

import (
	"errors"
	"github.com/gotid/god/lib/lang"
	"time"
)

// 表示一个超时错误。
var errTimeout = errors.New("timeout")

type (
	// Ticker 接口包装了 Chan 和 Stop 方法。
	Ticker interface {
		Chan() <-chan time.Time
		Stop()
	}

	// FakeTicker 接口用于单元测试。
	FakeTicker interface {
		Ticker
		Done()
		Tick()
		Wait(d time.Duration) error
	}

	fakeTicker struct {
		c    chan time.Time
		done chan lang.PlaceholderType
	}

	realTicker struct {
		*time.Ticker
	}
)

func (rt *realTicker) Chan() <-chan time.Time {
	return rt.C
}

func NewTicker(d time.Duration) Ticker {
	return &realTicker{
		Ticker: time.NewTicker(d),
	}
}

func (ft *fakeTicker) Chan() <-chan time.Time {
	return ft.c
}

func (ft *fakeTicker) Stop() {
	close(ft.c)
}

func (ft *fakeTicker) Done() {
	ft.done <- lang.Placeholder
}

func (ft *fakeTicker) Tick() {
	ft.c <- time.Now()
}

func (ft *fakeTicker) Wait(d time.Duration) error {
	select {
	case <-time.After(d):
		return errTimeout
	case <-ft.done:
		return nil
	}
}

func NewFakeTicker() FakeTicker {
	return &fakeTicker{
		c:    make(chan time.Time, 1),
		done: make(chan lang.PlaceholderType, 1),
	}
}
