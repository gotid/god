package threading

import (
	"bytes"
	"github.com/gotid/god/lib/rescue"
	"runtime"
	"strconv"
)

// RunSafe 执行函数 fn，若panic则记录日志。
func RunSafe(fn func()) {
	defer rescue.Recover()

	fn()
}

// GoSafe 用协程执行 fn，若panic则记录日志。
func GoSafe(fn func()) {
	go RunSafe(fn)
}

// RoutineId 仅用于调试，永远不要在生产环境使用。
func RoutineId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	// 有错，只返回0
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return n
}
