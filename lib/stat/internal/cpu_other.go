//go:build !linux
// +build !linux

package internal

func RefreshCpuUsage() uint64 {
	return 0
}
