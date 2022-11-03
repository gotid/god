package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/api/internal/response"
	"github.com/gotid/god/lib/breaker"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
)

const breakerSeparator = "://"

// BreakerHandler 返回一个自动熔断中间件。
func BreakerHandler(method, path string, metrics *stat.Metrics) func(handler http.Handler) http.Handler {
	brk := breaker.New(breaker.WithName(strings.Join([]string{method, path}, breakerSeparator)))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			promise, err := brk.Allow()
			if err != nil {
				metrics.AddDrop()
				logx.Errorf("[http] 丢弃，%s - %s - %s",
					r.RequestURI, httpx.GetRemoteAddr(r), r.UserAgent())
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			cw := &response.WithCodeResponseWriter{Writer: w}
			defer func() {
				if cw.Code < http.StatusInternalServerError {
					promise.Accept()
				} else {
					promise.Reject(fmt.Sprintf("%d %s", cw.Code, http.StatusText(cw.Code)))
				}
			}()

			next.ServeHTTP(cw, r)
		})
	}
}
