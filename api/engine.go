package api

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gotid/god/api/httpx"

	"github.com/gotid/god/api/handler"
	"github.com/gotid/god/api/internal"
	"github.com/gotid/god/api/router"
	"github.com/gotid/god/lib/codec"
	"github.com/gotid/god/lib/load"
	"github.com/gotid/god/lib/stat"
	"github.com/justinas/alice"
)

// 最高 CPU 用量。用 1000m 表示 cpu 负载为 100%。
const topCpuUsage = 1000

var ErrSignatureConfig = errors.New("错误的签名配置")

// engine 是一个 API 内部引擎。
type engine struct {
	conf                 ServerConf
	routes               []featuredRoutes
	middlewares          []Middleware
	unauthorizedCallback handler.UnauthorizedCallback
	unsignedCallback     handler.UnsignedCallback
	shedder              load.Shedder
	priorityShedder      load.Shedder
	tlsConfig            *tls.Config
}

// 返回一个新的 API 内部引擎。
func newEngine(c ServerConf) *engine {
	e := &engine{conf: c}

	// 启用 CPU 负载泄流阀
	if c.CpuThreshold > 0 {
		e.shedder = load.NewAdaptiveShedder(load.WithCpuThreshold(c.CpuThreshold))
		e.priorityShedder = load.NewAdaptiveShedder(load.WithCpuThreshold(
			(c.CpuThreshold + topCpuUsage) >> 1),
		)
	}

	return e
}

// addRoutes 增加一批特定路由
func (e *engine) addRoutes(r featuredRoutes) {
	e.routes = append(e.routes, r)
}

// setUnauthorizedCallback 设置未授权回调函数
func (e *engine) setUnauthorizedCallback(callback handler.UnauthorizedCallback) {
	e.unauthorizedCallback = callback
}

// setUnsignedCallback 设置未签名回调函数
func (e *engine) setUnsignedCallback(callback handler.UnsignedCallback) {
	e.unsignedCallback = callback
}

// Start 启动 API 引擎
func (e *engine) Start() error {
	return e.StartWithRouter(router.NewRouter())
}

// StartWithRouter 启动路由器
func (e *engine) StartWithRouter(router router.Router) error {
	if err := e.bindRoutes(router); err != nil {
		return err
	}

	if len(e.conf.CertFile) == 0 && len(e.conf.KeyFile) == 0 {
		return internal.StartHttp(e.conf.Host, e.conf.Port, router)
	}

	return internal.StartHttps(e.conf.Host, e.conf.Port, e.conf.CertFile, e.conf.KeyFile, router)
}

func (e *engine) bindRoutes(router router.Router) error {
	metrics := e.createMetrics()

	for _, route := range e.routes {
		if err := e.bindFeaturedRoutes(router, route, metrics); err != nil {
			return err
		}
	}

	return nil
}

func (e *engine) bindRoute(fr featuredRoutes, router router.Router, metrics *stat.Metrics,
	route Route, verifier func(chain alice.Chain) alice.Chain) error {
	chain := alice.New(
		handler.TraceHandler(e.conf.Name, route.Path), // 链路追踪
		e.getLogHandler(),                                          // 日志记录
		handler.PrometheusHandler(route.Path),                      // 请求时长和响应码监控
		handler.MaxConns(e.conf.MaxConns),                          // 并发限制
		handler.BreakerHandler(route.Method, route.Path, metrics),  // 自动熔断
		handler.ShedderHandler(e.getShedder(fr.priority), metrics), // 负载均衡
		handler.TimeoutHandler(e.checkedTimeout(fr.timeout)),       // 超时控制
		handler.RecoverHandler,                                     // 异常捕获
		handler.MetricHandler(metrics),                             // 耗时统计
		handler.MaxBytesHandler(e.conf.MaxBytes),                   // 内容长度限制
		handler.GzipHandler,                                        // Gzip压缩
	)
	chain = e.appendAuthHandler(fr, chain, verifier) // JWT 鉴权

	for _, middleware := range e.middlewares {
		chain = chain.Append(convertMiddleware(middleware))
	}
	handle := chain.ThenFunc(route.Handler)

	return router.Handle(route.Method, route.Path, handle)
}

// 创建 API 引擎统计指标。
func (e *engine) createMetrics() *stat.Metrics {
	var metrics *stat.Metrics

	if len(e.conf.Name) > 0 {
		metrics = stat.NewMetrics(e.conf.Name)
	} else {
		metrics = stat.NewMetrics(fmt.Sprintf("%s:%d", e.conf.Host, e.conf.Port))
	}

	return metrics
}

// 绑定带有特色功能的路由（负载优先级、jwt校验、签名校验等）
func (e *engine) bindFeaturedRoutes(router router.Router, fr featuredRoutes, metrics *stat.Metrics) error {
	verifier, err := e.signatureVerifier(fr.signature)
	if err != nil {
		return err
	}

	for _, route := range fr.routes {
		if err := e.bindRoute(fr, router, metrics, route, verifier); err != nil {
			return err
		}
	}

	return nil
}

func (e *engine) signatureVerifier(signature signatureSetting) (func(chain alice.Chain) alice.Chain, error) {
	if !signature.enabled {
		return func(chain alice.Chain) alice.Chain {
			return chain
		}, nil
	}

	if len(signature.PrivateKeys) == 0 {
		if signature.Strict {
			return nil, ErrSignatureConfig
		} else {
			return func(chain alice.Chain) alice.Chain {
				return chain
			}, nil
		}
	}

	decrypters := make(map[string]codec.RsaDecryptor)
	for _, key := range signature.PrivateKeys {
		fingerprint := key.Fingerprint
		file := key.KeyFile
		decrypter, err := codec.NewRsaDecryptor(file)
		if err != nil {
			return nil, err
		}

		decrypters[fingerprint] = decrypter
	}

	return func(chain alice.Chain) alice.Chain {
		if e.unsignedCallback != nil {
			return chain.Append(handler.ContentSecurityHandler(
				decrypters, signature.Expire, signature.Strict, e.unsignedCallback))
		} else {
			return chain.Append(handler.ContentSecurityHandler(
				decrypters, signature.Expire, signature.Strict))
		}
	}, nil
}

// 添加JWT鉴权中间件
func (e *engine) appendAuthHandler(fr featuredRoutes, chain alice.Chain,
	verifier func(chain alice.Chain) alice.Chain) alice.Chain {
	if fr.jwt.enabled {
		if len(fr.jwt.prevSecret) == 0 {
			chain = chain.Append(handler.Authorize(fr.jwt.secret,
				handler.WithUnauthorizedCallback(e.unauthorizedCallback)))
		} else {
			chain = chain.Append(handler.Authorize(fr.jwt.secret,
				handler.WithPrevSecret(fr.jwt.prevSecret),
				handler.WithUnauthorizedCallback(e.unauthorizedCallback)))
		}
	}

	return verifier(chain)
}

// 获取日志记录器
func (e *engine) getLogHandler() alice.Constructor {
	if e.conf.Verbose {
		return handler.DetailedLogHandler
	} else {
		return handler.LogHandler
	}
}

// 获取负载泄流阀
func (e *engine) getShedder(priority bool) load.Shedder {
	if priority && e.priorityShedder != nil {
		return e.priorityShedder
	}

	return e.shedder
}

func (e *engine) setTlsConfig(cfg *tls.Config) {
	e.tlsConfig = cfg
}

func (e *engine) start(router httpx.Router) error {
	if err := e.bindRoutes(router); err != nil {
		return err
	}

	if len(e.conf.CertFile) == 0 && len(e.conf.KeyFile) == 0 {
		return internal.StartHttp(e.conf.Host, e.conf.Port, router)
	}

	return internal.StartHttps(e.conf.Host, e.conf.Port, e.conf.CertFile,
		e.conf.KeyFile, router, func(s *http.Server) {
			if e.tlsConfig != nil {
				s.TLSConfig = e.tlsConfig
			}
		})
}

func (e *engine) use(middleware Middleware) {
	e.middlewares = append(e.middlewares, middleware)
}

func (e *engine) checkedTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}

	return time.Duration(e.conf.Timeout) * time.Millisecond
}

func convertMiddleware(middleware Middleware) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return middleware(next.ServeHTTP)
	}
}
