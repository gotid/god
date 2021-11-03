package api

import (
	"errors"
	"log"
	"net/http"

	"git.zc0901.com/go/god/api/handler"
	"git.zc0901.com/go/god/api/router"
	"git.zc0901.com/go/god/lib/logx"
)

type (
	// Server 是一个 HTTP 服务器。
	Server struct {
		engine *engine
		opts   runOptions
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
	if len(opts) > 1 {
		return nil, errors.New("只允许一个 RunOption")
	}

	if err := c.Setup(); err != nil {
		return nil, err
	}

	server := &Server{
		engine: newEngine(c),
		opts: runOptions{
			start: func(e *engine) error {
				return e.Start()
			},
		},
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
	s.engine.AddRoutes(r)
}

// AddRoute 添加指定路由至服务器。
func (s *Server) AddRoute(r Route, opts ...RouteOption) {
	s.AddRoutes([]Route{r}, opts...)
}

// Start 启动服务器。
//
// 默认启用平滑关闭，可通过 proc.SetTimeToForceQuit 自定义平滑关闭时间。
func (s *Server) Start() {
	handleError(s.opts.start(s.engine))
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
	rt := router.NewRouter()
	rt.SetNotFoundHandler(handler)
	return WithRouter(rt)
}

// WithNotAllowedHandler 返回一个资源不允许访问的运行选项。
func WithNotAllowedHandler(handler http.Handler) RunOption {
	rt := router.NewRouter()
	rt.SetNotAllowedHandler(handler)
	return WithRouter(rt)
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
		server.engine.SetUnauthorizedCallback(callback)
	}
}

// WithUnsignedCallback 返回一个未签名回调的运行选项。
func WithUnsignedCallback(callback handler.UnsignedCallback) RunOption {
	return func(server *Server) {
		server.engine.SetUnsignedCallback(callback)
	}
}

// WithRouter 返回使用指定路由器的运行选项。
func WithRouter(router router.Router) RunOption {
	return func(server *Server) {
		server.opts.start = func(e *engine) error {
			return e.StartWithRouter(router)
		}
	}
}

func validateSecret(secret string) {
	if len(secret) < 8 {
		panic("JWT 密钥长度不能小于8位")
	}
}

func handleError(err error) {
	if err == nil || err == http.ErrServerClosed {
		return
	}

	logx.Error(err)
	panic(err)
}
