//go:build linux || darwin

package proc

import (
	"fmt"
	"github.com/gotid/god/lib/logx"
	"os"
	"path"
	"runtime/pprof"
	"syscall"
	"time"
)

const (
	goroutineProfile = "goroutine"
	debugLevel       = 2
)

func dumpGoroutines() {
	command := path.Base(os.Args[0])
	pid := syscall.Getpid()
	dumpFile := path.Join(os.TempDir(), fmt.Sprintf("%s-%d-goroutines-%s.dump",
		command, pid, time.Now().Format(timeFormat)))

	logx.Infof("收到导出协程信号，打印协程档案至 %s", dumpFile)

	if f, err := os.Create(dumpFile); err != nil {
		logx.Errorf("导出协程档案失败，错误：%v", err)
	} else {
		defer f.Close()
		pprof.Lookup(goroutineProfile).WriteTo(f, debugLevel)
	}
}
