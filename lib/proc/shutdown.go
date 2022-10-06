//go:build linux || darwin
// +build linux darwin

package proc

import (
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/threading"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	// 我们的队列大多数阻塞5秒钟，所以此处硬编码为 5500 毫秒
	waitTime   = 5500 * time.Millisecond
	wrapUpTime = time.Second
)

var (
	shutdownListeners        = new(listenerManager)
	wrapUpListeners          = new(listenerManager)
	delayTimeBeforeForceQuit = waitTime
)

// AddShutdownListener 添加函数 fn 作为一个关闭监听器。
// 返回的函数可用于等待调用fn。
func AddShutdownListener(fn func()) (waitForCalled func()) {
	return shutdownListeners.addListener(fn)
}

// AddWrapUpListener 添加函数 fn 作为一个圆满结束的监听器。
// 返回的函数可用于等待调用fn。
func AddWrapUpListener(fn func()) (waitForCalled func()) {
	return wrapUpListeners.addListener(fn)
}

// SetTimeToForceQuit 设置强制退出前的毫秒等待时间。
func SetTimeToForceQuit(duration time.Duration) {
	delayTimeBeforeForceQuit = duration
}

func gracefulStop(signals chan os.Signal) {
	signal.Stop(signals)

	logx.Info("收到信号 SIGTERM，关闭中...")
	go wrapUpListeners.notifyListeners()

	time.Sleep(wrapUpTime)
	go shutdownListeners.notifyListeners()

	time.Sleep(delayTimeBeforeForceQuit - wrapUpTime)
	logx.Infof("%v 毫秒后依然或者，即将强制终止该进程...", delayTimeBeforeForceQuit)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}

type listenerManager struct {
	lock      sync.Mutex
	waitGroup sync.WaitGroup
	listeners []func()
}

func (lm *listenerManager) addListener(fn func()) (waitForCalled func()) {
	lm.waitGroup.Add(1)

	lm.lock.Lock()
	lm.listeners = append(lm.listeners, func() {
		defer lm.waitGroup.Done()
		fn()
	})
	lm.lock.Unlock()

	return func() {
		lm.waitGroup.Wait()
	}
}

func (lm *listenerManager) notifyListeners() {
	lm.lock.Lock()
	defer lm.lock.Unlock()

	group := threading.NewRoutineGroup()
	for _, listener := range lm.listeners {
		group.RunSafe(listener)
	}
	group.Wait()
}
