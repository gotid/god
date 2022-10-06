package prometheus

import (
	"fmt"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/threading"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
)

var (
	once    sync.Once
	enabled syncx.AtomicBool
)

// Enabled 返回 prometheus 是否已经启动。
func Enabled() bool {
	return enabled.True()
}

// StartAgent 启动一个 prometheus 监控系统的代理。
func StartAgent(c Config) {
	if len(c.Host) == 0 {
		return
	}

	once.Do(func() {
		enabled.Set(true)
		threading.GoSafe(func() {
			http.Handle(c.Path, promhttp.Handler())
			addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
			logx.Infof("启动 prometheus 代理 %s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				logx.Error(err)
			}
		})
	})
}
