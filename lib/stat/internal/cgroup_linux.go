package internal

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gotid/god/lib/iox"
	"github.com/gotid/god/lib/lang"
)

// cgroups 的全称是control groups。
// cgroups为每种可以控制的资源定义了一个子系统，如cpu、内存、设备等。
// 详见：https://tech.meituan.com/2015/03/31/cgroups.html

const cgroupDir = "/sys/fs/cgroup"

type cgroup struct {
	cgroups map[string]string
}

// currentCgroup 获取当前进程的cgroup
func currentCgroup() (*cgroup, error) {
	cgroupFile := fmt.Sprintf("/proc/%d/cgroup", os.Getpid())
	lines, err := iox.ReadTextLines(cgroupFile, iox.WithoutBlank())
	if err != nil {
		return nil, err
	}

	cgroups := make(map[string]string)
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("无效的cgroup行：%s", line)
		}

		// 只读取cpu开头的相关信息
		subSys := parts[1]
		if !strings.HasPrefix(subSys, "cpu") {
			continue
		}

		// 组装每一个cpu子系统的文件路径
		// https://man7.org/linux/man-pages/man7/cgroups.7.html
		// comma-separated list of controllers for cgroup version 1
		fields := strings.Split(subSys, ",")
		for _, val := range fields {
			cgroups[val] = path.Join(cgroupDir, val)
		}
	}

	return &cgroup{cgroups: cgroups}, nil
}

// acctUsageAllCpus 报告所有Cpu的总用量
func (c *cgroup) acctUsageAllCpus() (uint64, error) {
	data, err := iox.ReadText(path.Join(c.cgroups["cpuacct"], "cpuacct.usage"))
	if err != nil {
		return 0, err
	}

	return parseUint(data)
}

// acctUsagePerCpu 报告每一个Cpu的单个用量
func (c *cgroup) acctUsagePerCpu() ([]uint64, error) {
	data, err := iox.ReadText(path.Join(c.cgroups["cpuacct"], "cpuacct.usage_percpu"))
	if err != nil {
		return nil, err
	}

	var usage []uint64
	for _, v := range strings.Fields(data) {
		u, err := parseUint(v)
		if err != nil {
			return nil, err
		}

		usage = append(usage, u)
	}

	return usage, nil
}

func (c *cgroup) cpuQuotaUs() (int64, error) {
	data, err := iox.ReadText(path.Join(c.cgroups["cpu"], "cpu.cfs.quota_us"))
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(data, 10, 64)
}

func (c *cgroup) cpuPeriodUs() (uint64, error) {
	data, err := iox.ReadText(path.Join(c.cgroups["cpu"], "cpu.cfs_period_us"))
	if err != nil {
		return 0, err
	}

	return parseUint(data)
}

func (c *cgroup) cpus() ([]uint64, error) {
	data, err := iox.ReadText(path.Join(c.cgroups["cpuset"], "cpuset.cpus"))
	if err != nil {
		return nil, err
	}

	return parseUints(data)
}

func parseUint(s string) (uint64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		if err.(*strconv.NumError).Err == strconv.ErrRange {
			return 0, nil
		} else {
			return 0, fmt.Errorf("cgroup: 入参不是正确的整型字符串: %s", s)
		}
	} else {
		if v < 0 {
			return 0, nil
		} else {
			return uint64(v), nil
		}
	}
}

func parseUints(val string) ([]uint64, error) {
	if val == "" {
		return nil, nil
	}

	ints := make(map[uint64]lang.PlaceholderType)
	parts := strings.Split(val, ",")
	for _, part := range parts {
		if strings.Contains(part, "-") {
			fields := strings.SplitN(part, "-", 2)
			min, err := parseUint(fields[0])
			if err != nil {
				return nil, fmt.Errorf("cgroup: 入参不是正确的整型字符串: %s", fields[0])
			}

			max, err := parseUint(fields[1])
			if err != nil {
				return nil, fmt.Errorf("cgroup: 入参不是正确的整型字符串: %s", fields[1])
			}

			if max < min {
				return nil, fmt.Errorf("cgroup: bad int list format: %s", val)
			}

			for i := min; i < max; i++ {
				ints[i] = lang.Placeholder
			}
		} else {
			v, err := parseUint(val)
			if err != nil {
				return nil, err
			}

			ints[v] = lang.Placeholder
		}
	}

	var sets []uint64
	for k := range ints {
		sets = append(sets, k)
	}

	return sets, nil
}
