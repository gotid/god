package handler

import (
	"net/http"

	"github.com/gotid/god/api/internal"
)

// MaxBytesHandler 返回一个限制请求体长度的中间件。
func MaxBytesHandler(n int64) func(http.Handler) http.Handler {
	if n <= 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > n {
				internal.Errorf(r, "请求实体过大，限制为：%d，但接收：%d，拒绝码：%d",
					n, r.ContentLength, http.StatusRequestEntityTooLarge)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
