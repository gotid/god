package rescue

import "github.com/gotid/god/lib/logx"

// Recover 用于panic时执行一系列清理和日志记录。
// 用法：
// defer Recover(func() {})
func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		logx.ErrorStack(p)
	}
}
