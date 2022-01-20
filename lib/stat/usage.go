package stat

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat/internal"
	"github.com/gotid/god/lib/threading"
)

const (
	// 每250毫秒检测一次最近5秒的cpu负载是否超过0.95
	cpuRefreshInterval = 250 * time.Millisecond // 250毫秒采集一次
	allRefreshInterval = time.Minute            // 1分钟打印一次
	beta               = 0.95
)

var cpuUsage int64

func init() {
	go func() {
		cpuTicker := time.NewTicker(cpuRefreshInterval)
		defer cpuTicker.Stop()

		allTicker := time.NewTicker(allRefreshInterval)
		defer allTicker.Stop()

		for {
			select {
			case <-cpuTicker.C:
				threading.RunSafe(func() {
					curUsage := internal.RefreshCpuUsage()
					prevUsage := atomic.LoadInt64(&cpuUsage)
					// cpu占用率计算公式：cpu = cpuᵗ⁻¹ * beta + cpuᵗ * (1 - beta)
					usage := int64(float64(prevUsage)*beta + float64(curUsage)*(1-beta))
					atomic.StoreInt64(&cpuUsage, usage)
				})
			case <-allTicker.C:
				if logEnabled.True() {
					printUsage()
				}
			}
		}
	}()
}

// CpuUsage 当前 CPU 用量
func CpuUsage() int64 {
	return atomic.LoadInt64(&cpuUsage)
}

// bytesToMb 字节转Mb
func bytesToMb(b uint64) float32 {
	return float32(b) / 1024 / 1024
}

// printUsage 打印系统负载
func printUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logx.Statf("CPU: %dm, 内存: 分配=%.1fMi, 累计分配=%.1fMi, 系统=%.1fMi, GC次数=%d",
		CpuUsage(), bytesToMb(m.Alloc), bytesToMb(m.TotalAlloc), bytesToMb(m.Sys), m.NumGC)
}
