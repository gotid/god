package httpx

import "net/http"

// Router 接口定义了一个处理 http 请求和响应的路由器。
type Router interface {
	http.Handler
	// Mount 挂载给定方法、路径的 http 处理器，即加入到路由树中。
	Mount(method, path string, handler http.Handler) error
	// SetNotFoundHandler 设置接口找不到的 http 处理器。
	SetNotFoundHandler(handler http.Handler)
	// SetNotAllowedHandler 设置方法不允许的 http 处理器。
	SetNotAllowedHandler(handler http.Handler)
}
