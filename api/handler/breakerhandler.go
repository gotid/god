package handler

import (
	"fmt"
	"net/http"
	"strings"

	"git.zc0901.com/go/god/api/httpx"
	"git.zc0901.com/go/god/api/internal/security"
	"git.zc0901.com/go/god/lib/breaker"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/stat"
)

const breakerSeparator = "://"

// BreakerHandler 返回一个断路器中间件。
func BreakerHandler(method, path string, metrics *stat.Metrics) func(http.Handler) http.Handler {
	brk := breaker.NewBreaker(breaker.WithName(strings.Join([]string{method, path}, breakerSeparator)))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			promise, err := brk.Allow()
			if err != nil {
				// 断路器端口，增加丢弃请求
				metrics.AddDrop()
				logx.Errorf("[HTTP] 请求已丢弃，%s - %s - %s",
					r.RequestURI, httpx.GetRemoteAddr(r), r.UserAgent())
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			cw := &security.WithCodeResponseWriter{Writer: w}
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
