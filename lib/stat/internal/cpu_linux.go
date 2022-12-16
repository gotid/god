package internal

import (
	"errors"
	"fmt"
	"github.com/gotid/god/lib/iox"
	"github.com/gotid/god/lib/logx"
	"strings"
	"sync"
	"time"
)

const (
	cpuFields = 8
	cpuTicks  = 100
)

var (
	quota     float64
	cores     uint64
	preSystem uint64
	preTotal  uint64
	initOnce  sync.Once
)

// 若没有 /proc 则忽略 cpu 计算，如 wsl linux。
func initialize() {
	cpus, err := cpuSets()
	if err != nil {
		logx.Error(err)
		return
	}

	cores = uint64(len(cpus))
	quota = float64(len(cpus))
	cq, err := cpuQuota()
	if err == nil {
		if cq != -1 {
			period, err := cpuPeriod()
			if err != nil {
				logx.Error(err)
				return
			}

			limit := float64(cq) / float64(period)
			if limit < quota {
				quota = limit
			}
		}
	}

	preSystem, err = systemCpuUsage()
	if err != nil {
		logx.Error(err)
		return
	}

	preTotal, err = totalCpuUsage()
	if err != nil {
		logx.Error(err)
		return
	}
}

// RefreshCpu 返回cpu用量。
func RefreshCpu() uint64 {
	initOnce.Do(initialize)

	total, err := totalCpuUsage()
	if err != nil {
		return 0
	}

	system, err := systemCpuUsage()
	if err != nil {
		return 0
	}

	var usage uint64
	cpuDelta := total - preTotal
	systemDelta := system - preSystem
	if cpuDelta > 0 && systemDelta > 0 {
		usage = uint64(float64(cpuDelta*cores*1e3) / (float64(systemDelta) * quota))
	}
	preSystem = system
	preTotal = total

	return usage
}

func cpuSets() ([]uint64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return nil, err
	}

	return cg.cpus()
}

func cpuQuota() (int64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return 0, err
	}

	return cg.cpuQuotaUs()
}

func cpuPeriod() (uint64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return 0, err
	}

	return cg.cpuPeriodUs()
}

func systemCpuUsage() (uint64, error) {
	lines, err := iox.ReadTextLines("/proc/stat", iox.WithoutBlank())
	if err != nil {
		return 0, err
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			if len(fields) < cpuFields {
				return 0, fmt.Errorf("错误的 /proc/stat 格式")
			}

			var totalClockTicks uint64
			for _, i := range fields[1:cpuFields] {
				v, err := parseUint(i)
				if err != nil {
					return 0, err
				}

				totalClockTicks += v
			}

			return (totalClockTicks * uint64(time.Second)) / cpuTicks, nil
		}
	}

	return 0, errors.New("错误的 /proc/stat 格式")
}

func totalCpuUsage() (usage uint64, err error) {
	var cg cgroup
	if cg, err = currentCgroup(); err != nil {
		return
	}

	return cg.usageAllCpus()
}
