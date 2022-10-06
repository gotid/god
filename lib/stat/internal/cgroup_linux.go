package internal

import (
	"bufio"
	"fmt"
	"github.com/gotid/god/lib/iox"
	"github.com/gotid/god/lib/lang"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cgroupDir   = "/sys/fs/cgroup"
	cpuStatFile = cgroupDir + "/cpu.stat"
	cpusetFile  = cgroupDir + "/cpuset.cpus.effective"
)

var (
	isUnified     bool
	inUserNS      bool
	isUnifiedOnce sync.Once
	nsOnce        sync.Once
)

type cgroup interface {
	cpuQuotaUs() (int64, error)
	cpuPeriodUs() (uint64, error)
	cpus() ([]uint64, error)
	usageAllCpus() (uint64, error)
}

func currentCgroup() (cgroup, error) {
	if isCgroup2UnifiedMode() {
		return currentCgroupV2()
	}

	return currentCgroupV1()
}

func currentCgroupV1() (cgroup, error) {
	cgroupFile := fmt.Sprintf("/proc/%d/cgroup", os.Getpid())
	lines, err := iox.ReadTextLines(cgroupFile, iox.WithoutBlank())
	if err != nil {
		return nil, err
	}

	cgroups := make(map[string]string)
	for _, line := range lines {
		cols := strings.Split(line, ":")
		if len(cols) != 3 {
			return nil, fmt.Errorf("无效的 cgroup 行: %s", line)
		}

		subsystems := cols[1]
		//只读取cpu相关行
		if !strings.HasPrefix(subsystems, "cpu") {
			continue
		}

		fields := strings.Split(subsystems, ",")
		for _, field := range fields {
			cgroups[field] = path.Join(cgroupDir, field)
		}
	}

	return &cgroupV1{
		cgroups: cgroups,
	}, nil
}

func currentCgroupV2() (cgroup, error) {
	lines, err := iox.ReadTextLines(cpuStatFile, iox.WithoutBlank())
	if err != nil {
		return nil, err
	}

	cgroups := make(map[string]string)
	for _, line := range lines {
		cols := strings.Fields(line)
		if len(cols) != 2 {
			return nil, fmt.Errorf("无效的 cgroupV2 行：%s", line)
		}

		cgroups[cols[0]] = cols[1]
	}

	return &cgroupV2{
		cgroups: cgroups,
	}, nil
}

// 返回程序是否在 cgroup v2 统一模式下运行。
func isCgroup2UnifiedMode() bool {
	isUnifiedOnce.Do(func() {
		var st unix.Statfs_t
		err := unix.Statfs(cgroupDir, &st)
		if err != nil {
			// 如果在用户命名空间中，则忽略 not found 错误
			if os.IsNotExist(err) && runningInUserNS() {
				isUnified = false
				return
			}
			panic(fmt.Sprintf("无法 statfs cgroup root：%s", err))
		}
		isUnified = st.Type == unix.CGROUP2_SUPER_MAGIC
	})

	return isUnified
}

// 检测当前是否在用户命名空间中运行。
func runningInUserNS() bool {
	nsOnce.Do(func() {
		file, err := os.Open("/proc/self/uid_map")
		if err != nil {
			// 该内核文件尽在系统支持用户命名空间时才会存在
			return
		}
		defer file.Close()

		buf := bufio.NewReader(file)
		l, _, err := buf.ReadLine()
		if err != nil {
			return
		}

		line := string(l)
		var a, b, c int64
		fmt.Sscanf(line, "%d %d %d", &a, &b, &c)

		/*
			我们假设，如果我们有一个完整的范围(0-4294967295)，那么我们就在初始的用户命名空间。
		*/
		if a == 0 && b == 0 && c == 4294967295 {
			return
		}

		inUserNS = true
	})

	return inUserNS
}

type cgroupV1 struct {
	cgroups map[string]string
}

func (c *cgroupV1) cpuQuotaUs() (int64, error) {
	data, err := iox.ReadText(path.Join(c.cgroups["cpu"], "cpu.cfs_quota_us"))
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(data, 10, 64)
}

func (c *cgroupV1) cpuPeriodUs() (uint64, error) {
	data, err := iox.ReadText(path.Join(c.cgroups["cpu"], "cpu.cfs_period_us"))
	if err != nil {
		return 0, err
	}

	return parseUint(data)
}

func (c *cgroupV1) cpus() ([]uint64, error) {
	data, err := iox.ReadText(path.Join(c.cgroups["cpuset"], "cpuset.cpus"))
	if err != nil {
		return nil, err
	}

	return parseUints(data)
}

func (c *cgroupV1) usageAllCpus() (uint64, error) {
	data, err := iox.ReadText(path.Join(c.cgroups["cpuacct"], "cpuacct.usage"))
	if err != nil {
		return 0, err
	}

	return parseUint(data)
}

type cgroupV2 struct {
	cgroups map[string]string
}

func (c *cgroupV2) cpuQuotaUs() (int64, error) {
	data, err := iox.ReadText(path.Join(cgroupDir, "cpu.cfs_quota_us"))
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(data, 10, 64)
}
func (c *cgroupV2) cpuPeriodUs() (uint64, error) {
	data, err := iox.ReadText(path.Join(cgroupDir, "cpu.cfs_period_us"))
	if err != nil {
		return 0, err
	}

	return parseUint(data)
}
func (c *cgroupV2) cpus() ([]uint64, error) {
	data, err := iox.ReadText(cpusetFile)
	if err != nil {
		return nil, err
	}

	return parseUints(data)
}

func (c *cgroupV2) usageAllCpus() (uint64, error) {
	usec, err := parseUint(c.cgroups["usage_usec"])
	if err != nil {
		return 0, err
	}

	return usec * uint64(time.Microsecond), nil
}

func parseUints(val string) ([]uint64, error) {
	if val == "" {
		return nil, nil
	}

	ints := make(map[uint64]lang.PlaceholderType)
	cols := strings.Split(val, ",")
	for _, r := range cols {
		if strings.Contains(r, "-") {
			fields := strings.SplitN(r, "-", 2)
			min, err := parseUint(fields[0])
			if err != nil {
				return nil, fmt.Errorf("cgroup: 错误的 int 列表格式：%s", val)
			}

			max, err := parseUint(fields[1])
			if err != nil {
				return nil, fmt.Errorf("cgroup: 错误的 int 列表格式：%s", val)
			}

			if max < min {
				return nil, fmt.Errorf("cgroup: 错误的 int 列表格式：%s", val)
			}

			for i := min; i <= max; i++ {
				ints[i] = lang.Placeholder
			}
		} else {
			v, err := parseUint(r)
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

func parseUint(s string) (uint64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		if err.(*strconv.NumError).Err == strconv.ErrRange {
			return 0, nil
		}

		return 0, fmt.Errorf("cgroup: 错误的 int 格式: %s", s)
	}

	if v < 0 {
		return 0, nil
	}

	return uint64(v), nil
}
