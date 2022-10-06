package fs

import (
	"os"
	"syscall"
)

// CloseOnExec 确保在进程 fork 时关闭文件。
func CloseOnExec(file *os.File) {
	if file != nil {
		syscall.CloseOnExec(int(file.Fd()))
	}
}
