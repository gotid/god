package api

import "net/http"

const (
	allowOrigin      = "Access-Control-Allow-Origin"
	allOrigins       = "*"
	allowMethods     = "Access-Control-Allow-Methods"
	allowHeaders     = "Access-Control-Allow-Headers"
	allowCredentials = "Access-Control-Allow-Credentials"
	headers          = "x-requested-with,content-type"
	methods          = "GET,HEAD,POST,PATCH,PUT,DELETE"
)

// CorsHandler 返回一个跨域请求处理器。
// origins 源站仅可指定一个或全部。
func CorsHandler(origins ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(origins) > 0 {
			w.Header().Set(allowOrigin, origins[0])
		} else {
			w.Header().Set(allowOrigin, allOrigins)
		}

		w.Header().Set(allowMethods, methods)
		w.Header().Set(allowHeaders, headers)
		w.Header().Set(allowCredentials, "true")
		w.WriteHeader(http.StatusNoContent)
	})
}
