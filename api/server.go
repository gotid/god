package api

import (
	"crypto/tls"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gotid/god/api/chain"
	"github.com/gotid/god/api/handler"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/api/internal/cors"
	"github.com/gotid/god/api/router"
	"github.com/gotid/god/lib/logx"
)

type (
	// Server 是一个 http 服务器。
	Server struct {
		ng     *engine
		router httpx.Router
	}

	// Option 自定义 Server 的方法。
	Option func(*Server)
)

// MustNewServer 返回一个给定配置和运行选项的服务器，遇错退出。
// 注意，后面的 Option 会覆盖前面的。
func MustNewServer(c Config, opts ...Option) *Server {
	server, err := NewServer(c, opts...)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

// NewServer 返回一个给定配置和运行选项的服务器。
// 注意，后面的 Option 会覆盖前面的。
func NewServer(c Config, opts ...Option) (*Server, error) {
	if err := c.Setup(); err != nil {
		return nil, err
	}

	server := &Server{
		ng:     newEngine(c),
		router: router.NewRouter(),
	}

	opts = append([]Option{WithNotFoundHandler(nil)}, opts...)
	for _, opt := range opts {
		opt(server)
	}

	return server, nil
}

// AddRoutes 添加一组路由到服务器 Server 中。
func (s *Server) AddRoutes(rs []Route, opts ...RouteOption) {
	r := featuredRoutes{
		routes: rs,
	}
	for _, opt := range opts {
		opt(&r)
	}
	s.ng.addRoutes(r)
}

// AddRoute 添加一个路由到服务器 Server 中。
func (s *Server) AddRoute(r Route, opts ...RouteOption) {
	s.AddRoutes([]Route{r}, opts...)
}

// PrintRoutes 打印已添加的路由至标准输出。
func (s *Server) PrintRoutes() {
	s.ng.print()
}

// Routes 返回该服务器中注册的路由列表。
func (s *Server) Routes() []Route {
	var routes []Route

	for _, r := range s.ng.routes {
		routes = append(routes, r.routes...)
	}

	return routes
}

// Start 启动服务器 Server。
// 默认情况下启用正常关机。
// 可使用 proc.SetTimeToForceQuit 自定义关机期间的行为。
func (s *Server) Start() {
	handleError(s.ng.start(s.router))
}

// Stop 停止服务器 Server。
func (s *Server) Stop() {
	logx.Close()
}

// Use 添加给定的中间件到服务器 Server。
func (s *Server) Use(middleware Middleware) {
	s.ng.use(middleware)
}

// ToMiddleware 将给定的处理器转为中间件 Middleware。
func ToMiddleware(handler func(next http.Handler) http.Handler) Middleware {
	return func(handle http.HandlerFunc) http.HandlerFunc {
		return handler(handle).ServeHTTP
	}
}

// WithChain 使用给定的中间件链 chain.Chain 代替默认的。
// JWT 鉴权中间件和通过 srv.Use 添加的中间件将被附带过去。
func WithChain(chn chain.Chain) Option {
	return func(svr *Server) {
		svr.ng.chain = chn
	}
}

// WithCors 启用给定来源的 CORS，默认允许所有来源（*）。
func WithCors(origin ...string) Option {
	return func(server *Server) {
		server.router.SetNotAllowedHandler(cors.NotAllowedHandler(nil, origin...))
		server.router = newCorsRouter(server.router, nil, origin...)
	}
}

// WithCustomCors 启用给定来源的 CORS，默认允许所有来源（*）。
// fn 允许调用者自定义响应。
func WithCustomCors(middlewareFn func(header http.Header), notAllowedFn func(http.ResponseWriter),
	origin ...string) Option {
	return func(server *Server) {
		server.router.SetNotAllowedHandler(cors.NotAllowedHandler(notAllowedFn, origin...))
		server.router = newCorsRouter(server.router, middlewareFn, origin...)
	}
}

// WithJwt 使用给定的秘钥进行 Jwt 身份鉴权。
func WithJwt(secret string) RouteOption {
	return func(r *featuredRoutes) {
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
	}
}

// WithJwtTransition 启用新老秘钥过度的 Jwt 身份鉴权。
// 这意味着新旧秘钥会在一段时间内协同工作。
func WithJwtTransition(secret, prevSecret string) RouteOption {
	return func(r *featuredRoutes) {
		// 为何不验证 prevSecret，因其已被用过，就算它不符合我们的要求，我们也得允许过度。
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
		r.jwt.prevSecret = prevSecret
	}
}

// WithMaxBytes 自定义最大请求字节数。
func WithMaxBytes(maxBytes int64) RouteOption {
	return func(r *featuredRoutes) {
		r.maxBytes = maxBytes
	}
}

// WithMiddlewares 添加一组给定的中间件到一组给定的路由上。
func WithMiddlewares(ms []Middleware, rs ...Route) []Route {
	for i := len(ms) - 1; i >= 0; i-- {
		rs = WithMiddleware(ms[i], rs...)
	}
	return rs
}

// WithMiddleware 添加一个给定的中间件到一组给定的路由上。
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

// WithNotFoundHandler 自定义未找到处理器。
func WithNotFoundHandler(handler http.Handler) Option {
	return func(server *Server) {
		notFoundHandler := server.ng.notFoundHandler(handler)
		server.router.SetNotFoundHandler(notFoundHandler)
	}
}

// WithNotAllowedHandler 自定义不允许访问处理器。
func WithNotAllowedHandler(handler http.Handler) Option {
	return func(server *Server) {
		server.router.SetNotAllowedHandler(handler)
	}
}

// WithPrefix 添加组名作为路由路径的前缀。
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

// WithPriority 区分路由优先级。
func WithPriority() RouteOption {
	return func(r *featuredRoutes) {
		r.priority = true
	}
}

// WithRouter 自定义服务器的路由器。
func WithRouter(router httpx.Router) Option {
	return func(server *Server) {
		server.router = router
	}
}

// WithSignature 启用内容签名校验。
func WithSignature(signature SignatureConfig) RouteOption {
	return func(r *featuredRoutes) {
		r.signature.enabled = true
		r.signature.Strict = signature.Strict
		r.signature.Expire = signature.Expire
		r.signature.PrivateKeys = signature.PrivateKeys
	}
}

// WithTimeout 自定义超时时长。
func WithTimeout(timeout time.Duration) RouteOption {
	return func(r *featuredRoutes) {
		r.timeout = timeout
	}
}

// WithTLSConfig 设置 https 配置。
func WithTLSConfig(cfg *tls.Config) Option {
	return func(svr *Server) {
		svr.ng.setTlsConfig(cfg)
	}
}

// WithUnauthorizedCallback 设置未授权回调函数。
func WithUnauthorizedCallback(callback handler.UnauthorizedCallback) Option {
	return func(svr *Server) {
		svr.ng.setUnauthorizedCallback(callback)
	}
}

// WithUnsignedCallback 设置签名失败回调函数。
func WithUnsignedCallback(callback handler.UnsignedCallback) Option {
	return func(svr *Server) {
		svr.ng.setUnsignedCallback(callback)
	}
}

func handleError(err error) {
	// ErrServerClosed 意为服务器已被手动关闭。
	if err == nil || err == http.ErrServerClosed {
		return
	}

	logx.Error(err)
	panic(err)
}

// 检查秘钥长度不能小于8位
func validateSecret(secret string) {
	if len(secret) < 8 {
		panic("秘钥长度不能小于 8 位")
	}
}

type corsRouter struct {
	httpx.Router
	middleware Middleware
}

func newCorsRouter(router httpx.Router, headerFn func(http.Header), origins ...string) httpx.Router {
	return &corsRouter{
		Router:     router,
		middleware: cors.Middleware(headerFn, origins...),
	}
}

func (c *corsRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.middleware(c.Router.ServeHTTP)(w, r)
}
