//go:build linux
// +build linux

package stat

import (
	"flag"
	"fmt"
	"github.com/gotid/god/lib/executors"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/sysx"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	testEnv        = "test.v"
	clusterNameKey = "CLUSTER_NAME"
	timeFormat     = "2006-01-02 15:04:05"
)

var (
	reporter     = logx.Alert
	lock         sync.RWMutex
	lessExecutor = executors.NewLessExecutor(5 * time.Minute)
	dropped      int32
	clusterName  = proc.Env(clusterNameKey)
)

func init() {
	if flag.Lookup(testEnv) != nil {
		SetReporter(nil)
	}
}

// SetReporter 指定汇报人为 fn。
func SetReporter(fn func(string)) {
	lock.Lock()
	defer lock.Unlock()
	reporter = fn
}

// StatReport 汇报给定的消息。
func Report(msg string) {
	lock.RLock()
	fn := reporter
	lock.RUnlock()

	if fn != nil {
		reported := lessExecutor.DoOrDiscard(func() {
			var builder strings.Builder
			builder.WriteString(fmt.Sprintln(time.Now().Format(timeFormat)))
			if len(clusterName) > 0 {
				builder.WriteString(fmt.Sprintf("cluster: %s\n", clusterName))
			}
			builder.WriteString(fmt.Sprintf("host: %s\n", sysx.Hostname()))
			dp := atomic.SwapInt32(&dropped, 0)
			if dp > 0 {
				builder.WriteString(fmt.Sprintf("dropped: %d\n", dp))
			}
			builder.WriteString(strings.TrimSpace(msg))
			fn(builder.String())
		})
		if !reported {
			atomic.AddInt32(&dropped, 1)
		}
	}
}
