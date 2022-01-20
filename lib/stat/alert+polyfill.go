//go:build !linux
// +build !linux

package stat

// Report 上报统计
func Report(string) {
}

// SetReporter 设置上报函数
func SetReporter(func(string)) {
}
