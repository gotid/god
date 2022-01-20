// Package rescue 营救——恢复弥补包
package rescue

import "github.com/gotid/god/lib/logx"

// Recover 恢复弥补函数：执行一组清理函数并尝试输出错误堆栈
func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	// 处理 panic 日志
	if p := recover(); p != nil {
		logx.ErrorStack(p)
	}
}
