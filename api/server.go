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
	runOptions struct {
		start func(*engine) error
	}

	RunOption func(*Server)

	Server struct {
		engine *engine
		opts   runOptions
	}
)

func MustNewServer(c Conf, opts ...RunOption) *Server {
	server, err := NewServer(c, opts...)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

func NewServer(c Conf, opts ...RunOption) (*Server, error) {
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

func (s *Server) AddRoutes(rs []Route, opts ...RouteOption) {
	r := featuredRoutes{routes: rs}
	for _, opt := range opts {
		opt(&r)
	}
	s.engine.AddRoutes(r)
}

func (s *Server) AddRoute(r Route, opts ...RouteOption) {
	s.AddRoutes([]Route{r}, opts...)
}

func (s *Server) Start() {
	handleError(s.opts.start(s.engine))
}

func (s *Server) Stop() {
	logx.Close()
}

func (s *Server) Use(middleware Middleware) {
	s.engine.use(middleware)
}

// 转为中间件
func ToMiddleware(handler func(next http.Handler) http.Handler) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return handler(next).ServeHTTP
	}
}

// 生成带有JWT鉴权的路由选项
func WithJwt(secret string) RouteOption {
	return func(r *featuredRoutes) {
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
	}
}

// 生成带有JWT鉴权的路由选项，带有新老秘钥转变处理
func WithJwtTransition(secret, prevSecret string) RouteOption {
	return func(r *featuredRoutes) {
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
		r.jwt.prevSecret = prevSecret
	}
}

func WithMiddlewares(ms []Middleware, rs ...Route) []Route {
	for i := len(ms) - 1; i >= 0; i-- {
		rs = WithMiddleware(ms[i], rs...)
	}
	return rs
}

// 将一组路由应用上一个中间件
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

// WithNotFoundHandler 附加资源未找到处理器选项
func WithNotFoundHandler(handler http.Handler) RunOption {
	rt := router.NewRouter()
	rt.SetNotFoundHandler(handler)
	return WithRouter(rt)
}

// WithNotAllowedHandler 附加资源不允许访问处理器选项
func WithNotAllowedHandler(handler http.Handler) RunOption {
	rt := router.NewRouter()
	rt.SetNotAllowedHandler(handler)
	return WithRouter(rt)
}

// WithPriority 附加高优先级路由选项
func WithPriority() RouteOption {
	return func(r *featuredRoutes) {
		r.priority = true
	}
}

func WithSignature(signature SignatureConf) RouteOption {
	return func(r *featuredRoutes) {
		r.signature.enabled = true
		r.signature.Strict = signature.Strict
		r.signature.Expire = signature.Expire
		r.signature.PrivateKeys = signature.PrivateKeys
	}
}

func WithUnauthorizedCallback(callback handler.UnauthorizedCallback) RunOption {
	return func(server *Server) {
		server.engine.SetUnauthorizedCallback(callback)
	}
}

func WithUnsignedCallback(callback handler.UnsignedCallback) RunOption {
	return func(server *Server) {
		server.engine.SetUnsignedCallback(callback)
	}
}

func WithRouter(router router.Router) RunOption {
	return func(server *Server) {
		server.opts.start = func(e *engine) error {
			return e.StartWithRouter(router)
		}
	}
}

func validateSecret(secret string) {
	if len(secret) < 8 {
		panic("JWT秘钥长度不能小于8位")
	}
}

func handleError(err error) {
	if err == nil || err == http.ErrServerClosed {
		return
	}

	logx.Error(err)
	panic(err)
}
