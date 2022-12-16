package devserver

import (
	"encoding/json"
	"fmt"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/threading"
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/felixge/fgprof"
	"github.com/gotid/god/internal/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var once sync.Once

type Server struct {
	config *Config
	server *http.ServeMux
	routes []string
}

// StartAgent 根据配置，启动内部 http 服务器。
func StartAgent(c Config) {
	once.Do(func() {
		if c.Enable {
			s := NewServer(&c)
			s.StartAsync()
		}
	})
}

// NewServer 返回一个新的内部 http 服务器。
func NewServer(config *Config) *Server {
	return &Server{
		config: config,
		server: http.NewServeMux(),
	}
}

// StartAsync 启动内部 http 服务器后台。
func (s *Server) StartAsync() {
	s.addRoutes()
	threading.GoSafe(func() {
		addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
		logx.Infof("启动 http 开发服务器：%s", addr)
		if err := http.ListenAndServe(addr, s.server); err != nil {
			logx.Error(err)
		}
	})
}

func (s *Server) addRoutes() {
	// 将路由列表以 json 形式写入响应流
	s.handleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(s.routes)
	})

	// health
	s.handleFunc(s.config.HealthPath, health.CreateHttpHandler())

	// metrics
	if s.config.EnableMetrics {
		s.handleFunc(s.config.MetricsPath, promhttp.Handler().ServeHTTP)
	}

	// pprof
	if s.config.EnablePprof {
		s.handleFunc("/debug/fgprof", fgprof.Handler().(http.HandlerFunc))
		s.handleFunc("/debug/pprof", pprof.Index)
		s.handleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		s.handleFunc("/debug/pprof/profile", pprof.Profile)
		s.handleFunc("/debug/pprof/symbol", pprof.Symbol)
		s.handleFunc("/debug/pprof/trace", pprof.Trace)
	}
}

func (s *Server) handleFunc(pattern string, handler http.HandlerFunc) {
	s.server.HandleFunc(pattern, handler)
	s.routes = append(s.routes, pattern)
}
