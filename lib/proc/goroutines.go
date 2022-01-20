package proc

import (
	"fmt"
	"os"
	"path"
	"runtime/pprof"
	"syscall"
	"time"

	"git.zc0901.com/go/god/lib/logx"
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

	logx.Infof("收到 goroutine 转存信号，正在打印 goroutine 分析至 %s...", dumpFile)

	if f, err := os.Create(dumpFile); err != nil {
		logx.Errorf("转存 goroutine 分析失败，错误：%v", err)
	} else {
		defer f.Close()
		pprof.Lookup(goroutineProfile).WriteTo(f, debugLevel)
	}
}
