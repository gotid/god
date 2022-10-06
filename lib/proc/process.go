package proc

import (
	"os"
	"path/filepath"
)

var (
	procName string
	pid      int
)

func init() {
	procName = filepath.Base(os.Args[0])
	pid = os.Getpid()
}

// Pid 返回当前进程ID。
func Pid() int {
	return pid
}

// ProcessName 返回进程名称，与命令名称相同。
func ProcessName() string {
	return procName
}
