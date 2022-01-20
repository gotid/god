package handler

import (
	"compress/gzip"
	"net/http"
	"strings"

	"git.zc0901.com/go/god/api/httpx"
)

const gzipEncoding = "gzip"

// GzipHandler 返回一个通过 gzip 压缩 http 请求体的中间件。
func GzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get(httpx.ContentEncoding), gzipEncoding) {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = reader
		}

		next.ServeHTTP(w, r)
	})
}
