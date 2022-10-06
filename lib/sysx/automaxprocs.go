package sysx

import "go.uber.org/automaxprocs/maxprocs"

// 自动设置 GOMAXPROCS 以匹配 Linux 容器的 CPU 配额。
func init() {
	maxprocs.Set(maxprocs.Logger(nil))
}
