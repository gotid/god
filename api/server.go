package api

import (
	"crypto/tls"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gotid/god/api/httpx"

	"github.com/gotid/god/api/internal/cors"

	"github.com/gotid/god/api/handler"
	"github.com/gotid/god/api/router"
	"github.com/gotid/god/lib/logx"
)

type (
	// Server 是一个 HTTP 服务器。
	Server struct {
		engine *engine
		// opts   runOptions
		router httpx.Router
	}

	// RunOption 自定义服务器运行的函数。
	RunOption func(*Server)

	// 服务器运行的自定义项。
	runOptions struct {
		start func(*engine) error
	}
)

// MustNewServer 创建指定的服务器配置和运行选项的服务器。
//
// RunOption 选项稍后可被覆写。
//
// 创建出错，进程退出。
func MustNewServer(c ServerConf, opts ...RunOption) *Server {
	server, err := NewServer(c, opts...)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

// NewServer 创建指定的服务器配置和运行选项的服务器。
//
// RunOption 选项稍后可被覆写。
func NewServer(c ServerConf, opts ...RunOption) (*Server, error) {
	if err := c.Setup(); err != nil {
		return nil, err
	}

	server := &Server{
		engine: newEngine(c),
		router: router.NewRouter(),
	}

	for _, opt := range opts {
		opt(server)
	}

	return server, nil
}

// AddRoutes 添加一组指定路由至服务器。
func (s *Server) AddRoutes(rs []Route, opts ...RouteOption) {
	r := featuredRoutes{routes: rs}
	for _, opt := range opts {
		opt(&r)
	}
	s.engine.addRoutes(r)
}

// AddRoute 添加指定路由至服务器。
func (s *Server) AddRoute(r Route, opts ...RouteOption) {
	s.AddRoutes([]Route{r}, opts...)
}

// Start 启动服务器。
//
// 默认启用平滑关闭，可通过 proc.SetTimeToForceQuit 自定义平滑关闭时间。
func (s *Server) Start() {
	handleError(s.engine.start(s.router))
}

// Stop 关闭服务器。
func (s *Server) Stop() {
	logx.Close()
}

// Use 添加指定中间件至服务器。
func (s *Server) Use(middleware Middleware) {
	s.engine.use(middleware)
}

// ToMiddleware 将处理器转为中间件。
func ToMiddleware(handler func(next http.Handler) http.Handler) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return handler(next).ServeHTTP
	}
}

// WithCors 返回一个允许指定来源的CORS中间件，默认允许所有来源(*)。
func WithCors(origin ...string) RunOption {
	return func(server *Server) {
		server.router.SetNotAllowedHandler(cors.NotAllowedHandler(origin...))
		server.Use(cors.Middleware(origin...))
	}
}

// WithJwt 返回一个带有 JWT 鉴权的运行选项。
func WithJwt(secret string) RouteOption {
	return func(r *featuredRoutes) {
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
	}
}

// WithJwtTransition 返回一个兼容新老 JWT 密钥的运行选项。
func WithJwtTransition(secret, prevSecret string) RouteOption {
	return func(r *featuredRoutes) {
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
		r.jwt.prevSecret = prevSecret
	}
}

// WithMiddlewares 添加一组中间件至一组路由。
func WithMiddlewares(ms []Middleware, rs ...Route) []Route {
	for i := len(ms) - 1; i >= 0; i-- {
		rs = WithMiddleware(ms[i], rs...)
	}
	return rs
}

// WithMiddleware 添加一个中间件至一组路由，并返回该组路由。
func WithMiddleware(middleware Middleware, rs ...Route) []Route {
	routes := make([]Route, len(rs))

	for i := range rs {
		route := rs[i]
		routes[i] = Route{
			Method:  route.Method,
			Path:    route.Path,
			Handler: middleware(route.Handler),
		}
	}

	return routes
}

// WithNotFoundHandler 返回一个资源未找到运行选项。
func WithNotFoundHandler(handler http.Handler) RunOption {
	return func(s *Server) {
		s.router.SetNotFoundHandler(handler)
	}
}

// WithNotAllowedHandler 返回一个资源不允许访问的运行选项。
func WithNotAllowedHandler(handler http.Handler) RunOption {
	return func(s *Server) {
		s.router.SetNotAllowedHandler(handler)
	}
}

// WithPrefix 返回一个将 group 作为路由前缀的运行选项。
func WithPrefix(group string) RouteOption {
	return func(r *featuredRoutes) {
		var routes []Route
		for _, rt := range r.routes {
			p := path.Join(group, rt.Path)
			routes = append(routes, Route{
				Method:  rt.Method,
				Path:    p,
				Handler: rt.Handler,
			})
		}
		r.routes = routes
	}
}

// WithPriority 返回一个高优先级路由的运行选项。
func WithPriority() RouteOption {
	return func(r *featuredRoutes) {
		r.priority = true
	}
}

// WithSignature 返回一个签名校验的运行选项。
func WithSignature(signature SignatureConf) RouteOption {
	return func(r *featuredRoutes) {
		r.signature.enabled = true
		r.signature.Strict = signature.Strict
		r.signature.Expire = signature.Expire
		r.signature.PrivateKeys = signature.PrivateKeys
	}
}

// WithUnauthorizedCallback 返回一个未授权回调的运行选项。
func WithUnauthorizedCallback(callback handler.UnauthorizedCallback) RunOption {
	return func(server *Server) {
		server.engine.setUnauthorizedCallback(callback)
	}
}

// WithUnsignedCallback 返回一个未签名回调的运行选项。
func WithUnsignedCallback(callback handler.UnsignedCallback) RunOption {
	return func(server *Server) {
		server.engine.setUnsignedCallback(callback)
	}
}

// WithTimeout 返回一个指定超时时长的运行选项。
func WithTimeout(timeout time.Duration) RouteOption {
	return func(r *featuredRoutes) {
		r.timeout = timeout
	}
}

// WithTLSConfig 返回一个指定 tls 配置的运行选项。
func WithTLSConfig(cfg *tls.Config) RunOption {
	return func(s *Server) {
		s.engine.setTlsConfig(cfg)
	}
}

// WithRouter 返回使用指定路由器的运行选项。
func WithRouter(router router.Router) RunOption {
	return func(server *Server) {
		server.router = router
	}
}

func handleError(err error) {
	// ErrServerClosed 意味着服务器已被人为关闭。
	if err == nil || err == http.ErrServerClosed {
		return
	}

	logx.Error(err)
	panic(err)
}

func validateSecret(secret string) {
	if len(secret) < 8 {
		panic("JWT 密钥长度不能小于8位")
	}
}
