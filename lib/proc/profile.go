//go:build linux || darwin
// +build linux darwin

package proc

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gotid/god/lib/logx"
)

// DefaultMemProfileRate 默认内存分析速率。
// 参见 https://pkg.go.dev/runtime#pkg-variables
const DefaultMemProfileRate = 4096

// 若启动 profile，需为非零值。
var started uint32

// Profile 表示一个活动的分析会话。
type Profile struct {
	// 每个分析后的清理函数
	closers []func()

	// 记录分析是否已停止
	stopped uint32
}

func (p *Profile) close() {
	for _, closer := range p.closers {
		closer()
	}
}

func (p *Profile) startBlockProfile() {
	fn := createDumpFile("block")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("profile: could not create block profile %q: %v", fn, err)
		return
	}

	runtime.SetBlockProfileRate(1)
	logx.Infof("profile: block profiling enabled, %s", fn)
	p.closers = append(p.closers, func() {
		pprof.Lookup("block").WriteTo(f, 0)
		f.Close()
		runtime.SetBlockProfileRate(0)
		logx.Infof("profile: block profiling disabled, %s", fn)
	})
}

func (p *Profile) startCpuProfile() {
	fn := createDumpFile("cpu")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("profile: could not create cpu profile %q: %v", fn, err)
		return
	}

	logx.Infof("profile: cpu profiling enabled, %s", fn)
	pprof.StartCPUProfile(f)
	p.closers = append(p.closers, func() {
		pprof.StopCPUProfile()
		f.Close()
		logx.Infof("profile: cpu profiling disabled, %s", fn)
	})
}

func (p *Profile) startMemProfile() {
	fn := createDumpFile("mem")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("profile: could not create memory profile %q: %v", fn, err)
		return
	}

	old := runtime.MemProfileRate
	runtime.MemProfileRate = DefaultMemProfileRate
	logx.Infof("profile: memory profiling enabled (rate %d), %s", runtime.MemProfileRate, fn)
	p.closers = append(p.closers, func() {
		pprof.Lookup("heap").WriteTo(f, 0)
		f.Close()
		runtime.MemProfileRate = old
		logx.Infof("profile: memory profiling disabled, %s", fn)
	})
}

func (p *Profile) startMutexProfile() {
	fn := createDumpFile("mutex")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("profile: could not create mutex profile %q: %v", fn, err)
		return
	}

	runtime.SetMutexProfileFraction(1)
	logx.Infof("profile: mutex profiling enabled, %s", fn)
	p.closers = append(p.closers, func() {
		if mp := pprof.Lookup("mutex"); mp != nil {
			mp.WriteTo(f, 0)
		}
		f.Close()
		runtime.SetMutexProfileFraction(0)
		logx.Infof("profile: mutex profiling disabled, %s", fn)
	})
}

func (p *Profile) startThreadCreateProfile() {
	fn := createDumpFile("threadcreate")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("profile: could not create threadcreate profile %q: %v", fn, err)
		return
	}

	logx.Infof("profile: threadcreate profiling enabled, %s", fn)
	p.closers = append(p.closers, func() {
		if mp := pprof.Lookup("threadcreate"); mp != nil {
			mp.WriteTo(f, 0)
		}
		f.Close()
		logx.Infof("profile: threadcreate profiling disabled, %s", fn)
	})
}

func (p *Profile) startTraceProfile() {
	fn := createDumpFile("trace")
	f, err := os.Create(fn)
	if err != nil {
		logx.Errorf("profile: could not create trace output file %q: %v", fn, err)
		return
	}

	if err := trace.Start(f); err != nil {
		logx.Errorf("profile: could not start trace: %v", err)
		return
	}

	logx.Infof("profile: trace enabled, %s", fn)
	p.closers = append(p.closers, func() {
		trace.Stop()
		logx.Infof("profile: trace disabled, %s", fn)
	})
}

// Stop 停止会话并刷新任何未写入的数据。
func (p *Profile) Stop() {
	if !atomic.CompareAndSwapUint32(&p.stopped, 0, 1) {
		// someone has already called close
		return
	}
	p.close()
	atomic.StoreUint32(&started, 0)
}

// StartProfile 启动一个新的分析会话。
// 调用者应调用 Stop 方法以清理会话。
func StartProfile() Stopper {
	if !atomic.CompareAndSwapUint32(&started, 0, 1) {
		logx.Error("profile: Start() already called")
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

		logx.Info("profile: caught interrupt, stopping profiles")
		prof.Stop()

		signal.Reset()
		syscall.Kill(os.Getpid(), syscall.SIGINT)
	}()

	return &prof
}

func createDumpFile(kind string) string {
	command := path.Base(os.Args[0])
	pid := syscall.Getpid()
	return path.Join(os.TempDir(), fmt.Sprintf("%s-%d-%s-%s.pprof",
		command, pid, kind, time.Now().Format(timeFormat)))
}
