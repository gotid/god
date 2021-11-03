package proc

import (
	"os"
	"path/filepath"
)

var (
	procName string
	pid      int
)

// init 初始化当前程序进程名称和进程编号
func init() {
	procName = filepath.Base(os.Args[0])
	pid = os.Getpid()
}

// ProcessName 返回当前进程名称。
func ProcessName() string {
	return procName
}

// Pid 返回当前进程ID。
func Pid() int {
	return pid
}
