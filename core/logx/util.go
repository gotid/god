package logx

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// 根据调用深度返回文件路径:行号
func getCaller(callDepth int) string {
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		return ""
	}

	return prettyCaller(file, line)
}

func getTimestamp() string {
	return time.Now().Format(timeFormat)
}

// 返回文件路径:行号
func prettyCaller(file string, line int) string {
	// a.txt
	idx := strings.LastIndexByte(file, '/')
	if idx < 0 {
		return fmt.Sprintf("%s:%d", file, line)
	}

	// x/a.txt
	idx = strings.LastIndexByte(file[:idx], '/')
	if idx < 0 {
		return fmt.Sprintf("%s:%d", file, line)
	}

	// x/y/z/a.txt
	return fmt.Sprintf("%s:%d", file[idx+1:], line)
}
