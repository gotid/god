package rest

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gotid/god/lib/codec"
	"github.com/gotid/god/lib/load"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/rest/chain"
	"github.com/gotid/god/rest/handler"
	"github.com/gotid/god/rest/httpx"
	"github.com/gotid/god/rest/internal"
	"github.com/gotid/god/rest/internal/response"
	"net/http"
	"sort"
	"time"
)

// 使用 1000m 来表示 100%
const topCpuUsage = 1000

// ErrSignatureConfig 指示签名配置的错误。
var ErrSignatureConfig = errors.New("签名配置错误")

type engine struct {
	config               Config
	routes               []featuredRoutes
	unauthorizedCallback handler.UnauthorizedCallback
	unsignedCallback     handler.UnsignedCallback
	chain                chain.Chain
	middlewares          []Middleware
	shedder              load.Shedder
	priorityShedder      load.Shedder
	tlsConfig            *tls.Config
}

func newEngine(c Config) *engine {
	ng := &engine{
		config: c,
	}
	if c.CpuThreshold > 0 {
		ng.shedder = load.NewAdaptiveShedder(load.WithCpuThreshold(c.CpuThreshold))

		priorityThreshold := (c.CpuThreshold + topCpuUsage) >> 1
		ng.priorityShedder = load.NewAdaptiveShedder(load.WithCpuThreshold(priorityThreshold))
	}

	return ng
}

func (ng *engine) addRoutes(r featuredRoutes) {
	ng.routes = append(ng.routes, r)
}

func (ng *engine) appendAuthHandler(fr featuredRoutes, chn chain.Chain,
	verifier func(chain.Chain) chain.Chain) chain.Chain {
	if fr.jwt.enabled {
		if len(fr.jwt.prevSecret) == 0 {
			chn = chn.Append(handler.Authorize(fr.jwt.secret,
				handler.WithUnauthorizedCallback(ng.unauthorizedCallback)))
		} else {
			chn = chn.Append(handler.Authorize(fr.jwt.secret,
				handler.WithPrevSecret(fr.jwt.prevSecret),
				handler.WithUnauthorizedCallback(ng.unauthorizedCallback)))
		}
	}

	return verifier(chn)
}

func (ng *engine) bindFeaturedRoutes(router httpx.Router, fr featuredRoutes, metrics *stat.Metrics) error {
	verifier, err := ng.signatureVerifier(fr.signature)
	if err != nil {
		return err
	}

	for _, route := range fr.routes {
		if err := ng.bindRoute(fr, router, metrics, route, verifier); err != nil {
			return err
		}
	}

	return nil
}

func (ng *engine) bindRoute(fr featuredRoutes, router httpx.Router, metrics *stat.Metrics,
	route Route, verifier func(chain.Chain) chain.Chain) error {
	chn := ng.chain
	if chn == nil {
		chn = chain.New(
			handler.TracingHandler(ng.config.Name, route.Path),
			ng.getLogHandler(),
			handler.PrometheusHandler(route.Path),
			handler.MaxConns(ng.config.MaxConns),
			handler.BreakerHandler(route.Method, route.Path, metrics),
			handler.SheddingHandler(ng.getShedder(fr.priority), metrics),
			handler.TimeoutHandler(ng.checkedTimeout(fr.timeout)),
			handler.RecoverHandler,
			handler.MetricHandler(metrics),
			handler.MaxBytesHandler(ng.checkedMaxBytes(fr.maxBytes)),
			handler.GunzipHandler,
		)
	}

	chn = ng.appendAuthHandler(fr, chn, verifier)

	for _, middleware := range ng.middlewares {
		chn = chn.Append(convertMiddleware(middleware))
	}
	handle := chn.ThenFunc(route.Handler)

	return router.Handle(route.Method, route.Path, handle)
}

func (ng *engine) bindRoutes(router httpx.Router) error {
	metrics := ng.createMetrics()

	for _, fr := range ng.routes {
		if err := ng.bindFeaturedRoutes(router, fr, metrics); err != nil {
			return err
		}
	}

	return nil
}

func (ng *engine) checkedMaxBytes(bytes int64) int64 {
	if bytes > 0 {
		return bytes
	}

	return ng.config.MaxBytes
}

func (ng *engine) checkedTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}

	return time.Duration(ng.config.Timeout) * time.Millisecond
}

func (ng *engine) createMetrics() *stat.Metrics {
	var metrics *stat.Metrics

	if len(ng.config.Name) > 0 {
		metrics = stat.NewMetrics(ng.config.Name)
	} else {
		metrics = stat.NewMetrics(fmt.Sprintf("%s:%d", ng.config.Host, ng.config.Port))
	}

	return metrics
}

func (ng *engine) getLogHandler() func(http.Handler) http.Handler {
	if ng.config.Verbose {
		return handler.DetailedLogHandler
	}

	return handler.LogHandler
}

func (ng *engine) getShedder(priority bool) load.Shedder {
	if priority && ng.priorityShedder != nil {
		return ng.priorityShedder
	}

	return ng.shedder
}

// notFoundHandler returns a middleware that handles 404 not found requests.
func (ng *engine) notFoundHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chn := chain.New(
			handler.TracingHandler(ng.config.Name, ""),
			ng.getLogHandler(),
		)

		var h http.Handler
		if next != nil {
			h = chn.Then(next)
		} else {
			h = chn.Then(http.NotFoundHandler())
		}

		cw := response.NewHeaderOnceResponseWriter(w)
		h.ServeHTTP(cw, r)
		cw.WriteHeader(http.StatusNotFound)
	})
}

func (ng *engine) print() {
	var routes []string

	for _, fr := range ng.routes {
		for _, route := range fr.routes {
			routes = append(routes, fmt.Sprintf("%s %s", route.Method, route.Path))
		}
	}

	sort.Strings(routes)

	fmt.Println("路由：")
	for _, route := range routes {
		fmt.Printf("  %s\n", route)
	}
}

func (ng *engine) setTlsConfig(cfg *tls.Config) {
	ng.tlsConfig = cfg
}

func (ng *engine) setUnauthorizedCallback(callback handler.UnauthorizedCallback) {
	ng.unauthorizedCallback = callback
}

func (ng *engine) setUnsignedCallback(callback handler.UnsignedCallback) {
	ng.unsignedCallback = callback
}

func (ng *engine) signatureVerifier(signature signatureSetting) (func(chain.Chain) chain.Chain, error) {
	if !signature.enabled {
		return func(chn chain.Chain) chain.Chain {
			return chn
		}, nil
	}

	if len(signature.PrivateKeys) == 0 {
		if signature.Strict {
			return nil, ErrSignatureConfig
		}

		return func(chn chain.Chain) chain.Chain {
			return chn
		}, nil
	}

	decryptors := make(map[string]codec.RsaDecryptor)
	for _, key := range signature.PrivateKeys {
		fingerprint := key.Fingerprint
		file := key.KeyFile
		decryptor, err := codec.NewRsaDecryptor(file)
		if err != nil {
			return nil, err
		}

		decryptors[fingerprint] = decryptor
	}

	return func(chn chain.Chain) chain.Chain {
		if ng.unsignedCallback != nil {
			return chn.Append(handler.ContentSecurityHandler(
				decryptors, signature.Expire, signature.Strict, ng.unsignedCallback))
		}

		return chn.Append(handler.ContentSecurityHandler(decryptors, signature.Expire, signature.Strict))
	}, nil
}

func (ng *engine) start(router httpx.Router) error {
	if err := ng.bindRoutes(router); err != nil {
		return err
	}

	if len(ng.config.CertFile) == 0 && len(ng.config.KeyFile) == 0 {
		return internal.StartHttp(ng.config.Host, ng.config.Port, router, ng.withTimeout())
	}

	return internal.StartHttps(ng.config.Host, ng.config.Port, ng.config.CertFile,
		ng.config.KeyFile, router, func(svr *http.Server) {
			if ng.tlsConfig != nil {
				svr.TLSConfig = ng.tlsConfig
			}
		}, ng.withTimeout())
}

func (ng *engine) use(middleware Middleware) {
	ng.middlewares = append(ng.middlewares, middleware)
}

func (ng *engine) withTimeout() internal.StartOption {
	return func(svr *http.Server) {
		timeout := ng.config.Timeout
		if timeout > 0 {
			// 因子 0.8 是为了避免客户端发送的内容比实际更长，
			// 如果没有该超时设置，服务端将超时并响应 503 服务不可用，
			// 会触发断路器自动熔断。
			svr.ReadTimeout = 4 * time.Duration(timeout) * time.Millisecond / 5
			// 因子 0.9 是为了避免客户端未设置写超时的情况下不读取响应，服务器将超时并响应 503 服务不可用，
			// 会触发断路器自动熔断。
			svr.WriteTimeout = 9 * time.Duration(timeout) * time.Millisecond / 10
		}
	}
}

func convertMiddleware(ware Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return ware(next.ServeHTTP)
	}
}
