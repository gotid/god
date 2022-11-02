package rest

import (
	"net/http"
	"time"
)

type (
	// Middleware 定义一个中间件方法。
	Middleware func(next http.HandlerFunc) http.HandlerFunc

	// Route 是一个 http 路由。
	Route struct {
		Method  string
		Path    string
		Handler http.HandlerFunc
	}

	// RouteOption 定义特色路由的方法。
	RouteOption func(r *featuredRoutes)

	jwtSetting struct {
		enabled    bool
		secret     string
		prevSecret string
	}

	signatureSetting struct {
		SignatureConfig
		enabled bool
	}

	featuredRoutes struct {
		timeout   time.Duration
		priority  bool
		jwt       jwtSetting
		signature signatureSetting
		routes    []Route
		maxBytes  int64
	}
)
