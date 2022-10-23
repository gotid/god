package rpc

import (
	"github.com/gotid/god/lib/load"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/rpc/internal"
	"github.com/gotid/god/rpc/internal/auth"
	"github.com/gotid/god/rpc/internal/serverinterceptors"
	"google.golang.org/grpc"
	"log"
	"time"
)

// Server 是一个 rpc 服务器。
type Server struct {
	server   internal.Server
	register internal.RegisterFn
}

// MustNewServer 返回一个 rpc 服务器 Server，遇错退出。
func MustNewServer(c ServerConfig, register internal.RegisterFn) *Server {
	server, err := NewServer(c, register)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

// NewServer 返回一个 rpc 服务器 Server。
func NewServer(c ServerConfig, register internal.RegisterFn) (*Server, error) {
	var err error
	if err = c.Validate(); err != nil {
		return nil, err
	}

	var server internal.Server
	metrics := stat.NewMetrics(c.ListenOn)
	serverOptions := []internal.ServerOption{
		internal.WithMetrics(metrics),
		internal.WithHealth(c.Health),
	}

	if c.HasEtcd() {
		server, err = internal.NewPubServer(c.Etcd, c.ListenOn, serverOptions...)
		if err != nil {
			return nil, err
		}
	} else {
		server = internal.NewServer(c.ListenOn, serverOptions...)
	}

	server.SetName(c.Name)
	if err = setupInterceptors(server, c, metrics); err != nil {
		return nil, err
	}

	if err = c.Setup(); err != nil {
		return nil, err
	}

	svr := &Server{
		server:   server,
		register: register,
	}

	return svr, nil
}

// AddOptions 添加给定的选项。
func (s *Server) AddOptions(options ...grpc.ServerOption) {
	s.server.AddOptions(options...)
}

// AddStreamInterceptors 添加给定的流式拦截器。
func (s *Server) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	s.server.AddStreamInterceptors(interceptors...)
}

// AddUnaryInterceptors 添加给定的一元拦截器。
func (s *Server) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	s.server.AddUnaryInterceptors(interceptors...)
}

// Start 启动 rpc 服务器 Server。
// 默认启用正常关机。
// 使用 proc.SetTimeToForceQuit 可以自定义关机延迟时间。
func (s *Server) Start() {
	if err := s.server.Start(s.register); err != nil {
		logx.Error(err)
		panic(err)
	}
}

// Stop 停止 rpc 服务器 Server。
func (s *Server) Stop() {
	logx.Close()
}

// DontLogContentForMethod 禁用给定方法的日志内容。
func DontLogContentForMethod(method string) {
	serverinterceptors.DontLogContentForMethod(method)
}

// SetServerSlowThreshold 设置慢阈值。
func SetServerSlowThreshold(threshold time.Duration) {
	serverinterceptors.SetSlowThreshold(threshold)
}

func setupInterceptors(server internal.Server, c ServerConfig, metrics *stat.Metrics) error {
	if c.CpuThreshold > 0 {
		shedder := load.NewAdaptiveShedder(load.WithCpuThreshold(c.CpuThreshold))
		server.AddUnaryInterceptors(serverinterceptors.UnarySheddingInterceptor(shedder, metrics))
	}

	if c.Timeout > 0 {
		server.AddUnaryInterceptors(serverinterceptors.UnaryTimeoutInterceptor(
			time.Duration(c.Timeout) * time.Millisecond))
	}

	if c.Auth {
		authenticator, err := auth.NewAuthenticator(c.Redis.NewRedis(), c.Redis.Key, c.StrictControl)
		if err != nil {
			return err
		}

		server.AddStreamInterceptors(serverinterceptors.StreamAuthorizeInterceptor(authenticator))
		server.AddUnaryInterceptors(serverinterceptors.UnaryAuthorizeInterceptor(authenticator))
	}

	return nil
}
