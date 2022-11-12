//go:build !linux

package internal

// RefreshCpu 返回cpu用量，非linux系统返回0。
func RefreshCpu() uint64 {
	return 0
}
