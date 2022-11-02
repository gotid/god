package handler

import (
	"compress/gzip"
	"github.com/gotid/god/rest/httpx"
	"net/http"
	"strings"
)

const gzipEncoding = "gzip"

// GunzipHandler 返回一个解压缩 http 请求体的中间件。
func GunzipHandler(next http.Handler) http.Handler {
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
