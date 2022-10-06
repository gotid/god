//go:build linux || darwin
// +build linux darwin

package proc

import (
	"fmt"
	"github.com/gotid/god/lib/logx"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync/atomic"
	"syscall"
	"time"
)

// DefaultMemProfileRate 是默认的内存分析速率。
// 详见 http
const DefaultMemProfileRate = 4096

// 如果有一个分析在运行则为非零
var started uint32

// Profile 表示一个活跃中的分析会话。
type Profile struct {
	// 保存每个分析之后运行的清理函数
	closers []func()

	// 如果调用了 profile.Stop 则记录
	stopped uint32
}

func (p *Profile) Stop() {
	if !atomic.CompareAndSwapUint32(&p.stopped, 0, 1) {
		// 已调用过
		return
	}
	p.close()
	atomic.SwapUint32(&started, 0)
}

func (p *Profile) close() {
	for _, closer := range p.closers {
		closer()
	}
}

func (p *Profile) startCpuProfile() {
	fn := createDumpFile("cpu")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("分析：无法创建 cpu 分析文件 %q: %v", fn, err)
		return
	}

	logx.Infof("分析：cpu 分析已启用，%s", fn)
	pprof.StartCPUProfile(f)
	p.closers = append(p.closers, func() {
		pprof.StopCPUProfile()
		f.Close()
		logx.Infof("分析：cpu 分析已禁用，%s", fn)
	})
}

func (p *Profile) startMemProfile() {
	fn := createDumpFile("mem")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("分析：无法创建内存分析文件 %q: %v", fn, err)
		return
	}

	old := runtime.MemProfileRate
	runtime.MemProfileRate = DefaultMemProfileRate
	logx.Infof("分析：内存分析已启用（速率 %d），%s", runtime.MemProfileRate, fn)
	p.closers = append(p.closers, func() {
		pprof.Lookup("heap").WriteTo(f, 0)
		f.Close()
		runtime.MemProfileRate = old
		logx.Infof("分析：内存分析已禁用，%s", fn)
	})
}

func (p *Profile) startMutexProfile() {
	fn := createDumpFile("mutex")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("分析：无法创建互斥锁分析文件 %q: %v", fn, err)
		return
	}

	runtime.SetMutexProfileFraction(1)
	logx.Infof("分析：互斥锁分析已启用，%s", fn)
	p.closers = append(p.closers, func() {
		if mp := pprof.Lookup("mutex"); mp != nil {
			mp.WriteTo(f, 0)
		}
		f.Close()
		runtime.SetMutexProfileFraction(0)
		logx.Infof("分析：互斥锁分析已禁用，%s", fn)
	})
}

func (p *Profile) startBlockProfile() {
	fn := createDumpFile("block")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("分析：无法创建阻塞分析文件 %q: %v", fn, err)
		return
	}
	runtime.SetBlockProfileRate(1)
	logx.Infof("分析：阻塞分析已启用，%s", fn)
	p.closers = append(p.closers, func() {
		pprof.Lookup("block").WriteTo(f, 0)
		f.Close()
		runtime.SetBlockProfileRate(0)
		logx.Infof("分析：阻塞分析已禁用，%s", fn)
	})
}

func (p *Profile) startTraceProfile() {
	fn := createDumpFile("trace")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("分析：无法创建跟踪输出文件 %q: %v", fn, err)
	}

	if err := trace.Start(f); err != nil {
		logx.Errorf("分析：无法启动跟踪：%v", err)
		return
	}

	logx.Infof("分析：跟踪已启用，%s", fn)
	p.closers = append(p.closers, func() {
		trace.Stop()
		logx.Infof("分析：跟踪已禁用，%s", fn)
	})
}

func (p *Profile) startThreadCreateProfile() {
	fn := createDumpFile("threadcreate")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("分析：无法创建 threadcreate 分析文件 %q: %v", fn, err)
		return
	}

	logx.Infof("分析：threadcreate 分析已启用，%s", fn)
	p.closers = append(p.closers, func() {
		if mp := pprof.Lookup("threadcreate"); mp != nil {
			mp.WriteTo(f, 0)
		}
		f.Close()
		logx.Infof("分析：threadcreate 分析已禁用，%s", fn)
	})
}

func createDumpFile(kind string) string {
	command := path.Base(os.Args[0])
	pid := syscall.Getpid()
	return path.Join(os.TempDir(), fmt.Sprintf("%s-%d-%s-%s.pprof",
		command, pid, kind, time.Now().Format(timeFormat)))
}

// StartProfile 启动一个新的分析会话。
// 调用者应该调用返回的 Stop 方法进行清洗。
func StartProfile() Stopper {
	if !atomic.CompareAndSwapUint32(&started, 0, 1) {
		logx.Error("分析：已调用过 Start()")
		return nopStopper
	}

	var prof Profile
	prof.startCpuProfile()
	prof.startMemProfile()
	prof.startMutexProfile()
	prof.startBlockProfile()
	prof.startTraceProfile()
	prof.startThreadCreateProfile()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		<-c

		logx.Info("分析：拦截到中断，捕获停止中")
		prof.Stop()

		signal.Reset()
		syscall.Kill(os.Getpid(), syscall.SIGINT)
	}()

	return &prof
}
