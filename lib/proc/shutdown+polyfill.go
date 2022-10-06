//go:build windows
// +build windows

package proc

import "time"

// AddShutdownListener 添加函数 fn 作为一个关闭监听器。
// 返回的函数可用于等待调用fn。
func AddShutdownListener(fn func()) (waitForCalled func()) {
	return fn
}

// AddWrapUpListener 添加函数 fn 作为一个圆满结束的监听器。
// 返回的函数可用于等待调用fn。
func AddWrapListener(fn func()) (waitForCalled func()) {
	return fn
}

// SetTimeToForceQuit 设置强制退出前的毫秒等待时间。
func SetTimeToForceQuit(duration time.Duration) {

}
