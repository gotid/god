//go:build linux || darwin
// +build linux darwin

package proc

import (
	"github.com/gotid/god/lib/logx"
	"os"
	"os/signal"
	"syscall"
)

const timeFormat = "0102150405"

var done = make(chan struct{})

func init() {
	go func() {
		var profiler Stopper

		// https://golang.org/pkg/os/signal/#Notify
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGTERM)

		for {
			v := <-signals
			switch v {
			case syscall.SIGUSR1:
				dumpGoroutines()
			case syscall.SIGUSR2:
				if profiler == nil {
					profiler = StartProfile()
				} else {
					profiler.Stop()
					profiler = nil
				}
			case syscall.SIGTERM:
				select {
				case <-done:
				// 已经关闭
				default:
					close(done)
				}

				gracefulStop(signals)
			default:
				logx.Error("收到未注册的信号：", v)
			}
		}
	}()
}

// Done 返回通知进程退出的通道。
func Done() <-chan struct{} {
	return done
}
