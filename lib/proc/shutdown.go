//go:build linux || darwin
// +build linux darwin

package proc

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gotid/god/lib/logx"
)

const (
	wrapUpTime = time.Second
	// 使用 5500 毫秒，是因为我们的大多数队列都是阻塞 5秒。
	waitTime = 5500 * time.Millisecond
)

var (
	shutdownListeners        = new(ListenerManager)
	wrapUpListeners          = new(ListenerManager)
	delayTimeBeforeForceQuit = waitTime
)

// AddShutdownListener 添加一个程序关闭监听器
func AddShutdownListener(listener func()) (waitForCalled func()) {
	return shutdownListeners.add(listener)
}

// AddWrapUpListener 添加一个包装监听器
func AddWrapUpListener(listener func()) (waitForCalled func()) {
	return wrapUpListeners.add(listener)
}

func SetTimeToForceQuit(delay time.Duration) {
	delayTimeBeforeForceQuit = delay
}

// gracefulStop 平滑停止程序（为关闭类监听器的执行留有时间）
func gracefulStop(signals chan os.Signal) {
	signal.Stop(signals)

	logx.Info("捕获信号 SIGTERM，关闭中...")
	wrapUpListeners.notify()

	// 通知关闭类监听器，如定时执行器在关闭前进行任务 Flush
	time.Sleep(wrapUpTime)
	shutdownListeners.notify()

	// 静候5秒，等待监听器处理完毕，然后关闭程序
	time.Sleep(delayTimeBeforeForceQuit - wrapUpTime)
	logx.Infof("等 %v 秒后，将强杀该进程...", delayTimeBeforeForceQuit)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
