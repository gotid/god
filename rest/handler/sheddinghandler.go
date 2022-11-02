package handler

import (
	"github.com/gotid/god/lib/load"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/rest/httpx"
	"github.com/gotid/god/rest/internal/response"
	"net/http"
	"sync"
)

const serviceType = "api"

var (
	sheddingStat *load.SheddingStat
	lock         sync.Mutex
)

// SheddingHandler 返回一个自动降载中间件。
func SheddingHandler(shedder load.Shedder, metrics *stat.Metrics) func(http.Handler) http.Handler {
	if shedder == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	ensureSheddingStat()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sheddingStat.IncrTotal()
			promise, err := shedder.Allow()
			if err != nil {
				metrics.AddDrop()
				sheddingStat.IncrDrop()
				logx.Errorf("[http] 丢弃，%s - %s - %s",
					r.RequestURI, httpx.GetRemoteAddr(r), r.UserAgent())
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			cw := &response.WithCodeResponseWriter{Writer: w}
			defer func() {
				if cw.Code == http.StatusServiceUnavailable {
					promise.Fail()
				} else {
					sheddingStat.IncrPass()
					promise.Pass()
				}
			}()
			next.ServeHTTP(cw, r)
		})
	}
}

func ensureSheddingStat() {
	lock.Lock()
	if sheddingStat == nil {
		sheddingStat = load.NewSheddingStat(serviceType)
	}
	lock.Unlock()
}
