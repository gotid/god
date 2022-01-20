package handler

import (
	"net/http"
	"sync"

	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/api/internal/security"
	"github.com/gotid/god/lib/load"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
)

const serviceType = "API"

var (
	shedderStat *load.ShedderStat
	lock        sync.Mutex
)

// ShedderHandler 返回一个负载泄流阀中间件。
func ShedderHandler(shedder load.Shedder, metrics *stat.Metrics) func(http.Handler) http.Handler {
	if shedder == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	ensureShedderStat()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			shedderStat.IncrTotal()
			promise, err := shedder.Allow()
			if err != nil {
				metrics.AddDrop()
				shedderStat.IncrDrop()
				logx.Errorf("[HTTP] 请求被丢弃，%s - %s - %s",
					r.RequestURI, httpx.GetRemoteAddr(r), r.UserAgent())
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			cw := &security.WithCodeResponseWriter{Writer: w}
			defer func() {
				if cw.Code == http.StatusServiceUnavailable {
					promise.Fail()
				} else {
					shedderStat.IncrPass()
					promise.Pass()
				}
			}()

			next.ServeHTTP(cw, r)
		})
	}
}

func ensureShedderStat() {
	lock.Lock()
	if shedderStat == nil {
		shedderStat = load.NewShedderStat(serviceType)
	}
	lock.Unlock()
}
