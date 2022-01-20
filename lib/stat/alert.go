//go:build linux
// +build linux

package stat

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/executors"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/sysx"
	"github.com/gotid/god/lib/timex"
)

const (
	testEnv        = "test.v"
	clusterNameKey = "CLUSTER_NAME"
	timeFormat     = "2006-01-02 15:04:05"
)

var (
	lock                sync.RWMutex
	reporter            = logx.Alert
	shortTimeDispatcher = executors.NewShortTimeExecutor(5 * time.Minute)
	clusterName         = proc.Env(clusterNameKey)
	dropped             int32
)

func init() {
	if flag.Lookup(testEnv) != nil {
		SetReporter(nil)
	}
}

func Report(msg string) {
	lock.RLock()
	fn := reporter
	lock.RUnlock()

	if fn != nil {
		reported := shortTimeDispatcher.DoOrDiscard(func() {
			var b strings.Builder
			fmt.Fprintf(&b, "%s\n", timex.Time().Format(timeFormat))
			if len(clusterName) > 0 {
				fmt.Fprintf(&b, "集群名称: %s\n", clusterName)
			}
			fmt.Fprintf(&b, "主机: %s\n", sysx.Hostname())
			dp := atomic.SwapInt32(&dropped, 0)
			if dp > 0 {
				fmt.Fprintf(&b, "丢弃：%d\n", dp)
			}
			b.WriteString(strings.TrimSpace(msg))
			fn(b.String())
		})
		if !reported {
			atomic.AddInt32(&dropped, 1)
		}
	}
}

func SetReporter(fn func(string)) {
	lock.Lock()
	defer lock.Unlock()
	reporter = fn
}
