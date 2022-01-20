package prometheus

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/threading"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	once    sync.Once
	enabled syncx.AtomicBool
)

// Enabled 返回普罗米修斯是否启用。
func Enabled() bool {
	return enabled.True()
}

// StartAgent 启动普罗米修斯Http代理服务
func StartAgent(c Config) {
	once.Do(func() {
		// 未配置主机，表明不开启prometheus监控，直接返回
		if len(c.Host) == 0 {
			return
		}

		// 监听端口，等待Prometheus服务器的调用
		enabled.Set(true)
		threading.GoSafe(func() {
			http.Handle(c.Path, promhttp.Handler())
			addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
			logx.Infof("启动普罗米修斯代理，端口：%s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				logx.Error(err)
			}
		})
	})
}
