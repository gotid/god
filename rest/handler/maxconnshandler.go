package handler

import (
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/rest/internal"
	"net/http"
)

// MaxConns 返回一个并发控制中间件。
func MaxConns(n int) func(http.Handler) http.Handler {
	if n <= 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		latch := syncx.NewLimit(n)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if latch.TryBorrow() {
				defer func() {
					if err := latch.Return(); err != nil {
						logx.Error(err)
					}
				}()

				next.ServeHTTP(w, r)
			} else {
				internal.Errorf(r, "并发连接超过 %d，拒绝状态码 %d",
					n, http.StatusServiceUnavailable)
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		})
	}
}
