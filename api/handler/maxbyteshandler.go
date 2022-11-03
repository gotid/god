package handler

import (
	"github.com/gotid/god/api/internal"
	"net/http"
)

// MaxBytesHandler 返回一个限制请求体读取的中间件。
func MaxBytesHandler(n int64) func(http.Handler) http.Handler {
	if n <= 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > n {
				internal.Errorf(r, "请求体太大，限制为 %d，但收到 %d，拒绝状态码 %d",
					n, r.ContentLength, http.StatusRequestEntityTooLarge)
				w.WriteHeader(http.StatusRequestEntityTooLarge)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
