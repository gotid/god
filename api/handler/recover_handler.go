package handler

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"git.zc0901.com/go/god/api/internal"
)

// RecoverHandler API 异常捕获中间件
func RecoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if result := recover(); result != nil {
				internal.Error(r, fmt.Sprintf("%v\n%s", result, debug.Stack()))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
