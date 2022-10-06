package stat

import (
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat/internal"
	"github.com/gotid/god/lib/threading"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	// 每250毫秒计算一次过去5秒的CPU平均负载
	cpuRefreshInterval = 250 * time.Millisecond
	// 每1分钟进行一次全部统计
	allRefreshInterval = 1 * time.Minute
	// 移动的beta平均超参（上一次统计权重占95%）
	beta = 0.95
)

var cpuUsage int64

// 初始化cpu使用率
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
					curUsage := internal.RefreshCpu()
					preUsage := atomic.LoadInt64(&cpuUsage)
					// cpu = cpuᵗ⁻¹ * beta + cpuᵗ * (1 - beta)
					usage := int64(float64(preUsage)*beta + float64(curUsage)*(1-beta))
					atomic.SwapInt64(&cpuUsage, usage)
				})
			case <-allTicker.C:
				if logEnabled.True() {
					printUsage()
				}
			}
		}
	}()
}

// CpuUsage 返回上一次统计的 cpu 使用率
func CpuUsage() int64 {
	return atomic.LoadInt64(&cpuUsage)
}

func printUsage() {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	logx.Statf("CPU: %dm，内存: 当前=%.1fMi，总计=%.1fMi，系统=%.1fMi，GC=%d",
		CpuUsage(), bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
}

func bToMb(b uint64) float32 {
	return float32(b) / 1024 / 1024
}
