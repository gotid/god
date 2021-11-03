package api

import "net/http"

type (
	// Route 是一个 HTTP 路由。
	Route struct {
		Method  string           // 路由方法
		Path    string           // 路由路径
		Handler http.HandlerFunc // 路由处理器
	}

	// RouteOption 是一个路由自定义函数。
	RouteOption func(r *featuredRoutes)

	// Middleware 中间件是一个接收处理函数并返回新处理函数的函数。
	Middleware func(next http.HandlerFunc) http.HandlerFunc

	// 是一个 Json Web Token 设置结构体。
	jwtSetting struct {
		enabled    bool   // 是否启用jwt验证
		secret     string // jwt秘钥
		prevSecret string // 上一个jwt秘钥
	}

	// 签名设置
	signatureSetting struct {
		SignatureConf
		enabled bool // 是否启用签名校验
	}

	// 特色路由，支持高优先级、jwt令牌校验、签名校验
	featuredRoutes struct {
		priority  bool             // 带有高优先级的路由
		jwt       jwtSetting       // JWT 鉴权
		signature signatureSetting // 签名校验
		routes    []Route
	}
)
