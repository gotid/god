//go:build !linux
// +build !linux

package stat

// SetReporter 指定汇报人，如 logx.Alert。
func SetReporter(func(string)) {}

// Report 汇报给定的消息。
func Report(string) {}
