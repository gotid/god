package rescue

import "github.com/gotid/god/lib/logx"

// Recover 用于 panic 清理和记录。
// 用法：
//
// defer Recover(func() {})
func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		logx.ErrorStack(p)
	}
}
