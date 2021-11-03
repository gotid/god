package httpx

import "net/http"

// Router 表示一个处理 HTTP 请求的路由器。
type Router interface {
	http.Handler
	Handle(method, path string, handler http.Handler) error
	SetNotFoundHandler(handler http.Handler)
	SetNotAllowedHandler(handler http.Handler)
}
