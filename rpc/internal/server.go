package internal

import (
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/rpc/internal/serverinterceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
)

type (
	// ServerOption 自定义 serverOptions 的方法。
	ServerOption func(options *serverOptions)

	serverOptions struct {
		metrics *stat.Metrics
		health  bool
	}

	server struct {
		name string
		*baseServer
	}
)

// NewServer 返回一个 rpc 服务器 Server。
func NewServer(address string, opts ...ServerOption) Server {
	var options serverOptions
	for _, opt := range opts {
		opt(&options)
	}
	if options.metrics == nil {
		options.metrics = stat.NewMetrics(address)
	}

	return &server{
		baseServer: newBaseServer(address, &options),
	}
}

func (s *server) SetName(name string) {
	s.name = name
	s.baseServer.SetName(name)
}

func (s *server) Start(register RegisterFn) error {
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		serverinterceptors.UnaryTracingInterceptor, // 链路跟踪
		serverinterceptors.UnaryCrashInterceptor,
		serverinterceptors.UnaryStatInterceptor(s.metrics),
		serverinterceptors.UnaryPrometheusInterceptor, // 数据统计
		serverinterceptors.UnaryBreakerInterceptor,    // 自动熔断
	}
	unaryInterceptors = append(unaryInterceptors, s.unaryInterceptors...)

	streamInterceptors := []grpc.StreamServerInterceptor{
		serverinterceptors.StreamCrashInterceptor,
		serverinterceptors.StreamCrashInterceptor,
		serverinterceptors.StreamBreakerInterceptor, // 自动熔断
	}
	streamInterceptors = append(streamInterceptors, s.streamInterceptors...)

	options := append(s.options, WithUnaryServerInterceptors(unaryInterceptors...),
		WithStreamServerInterceptors(streamInterceptors...))
	svr := grpc.NewServer(options...)
	register(svr)

	// 注册 grpc 健康检查服务
	if s.health != nil {
		grpc_health_v1.RegisterHealthServer(svr, s.health)
		s.health.Resume()
	}

	// 确保关闭健康检查服务器和 grpc 服务器
	waitForCalled := proc.AddWrapUpListener(func() {
		if s.health != nil {
			s.health.Shutdown()
		}
		svr.GracefulStop()
	})
	defer waitForCalled()

	return svr.Serve(listen)
}

// WithMetrics 设置 grpc 服务器 Server 的统计指标。
func WithMetrics(metrics *stat.Metrics) ServerOption {
	return func(options *serverOptions) {
		options.metrics = metrics
	}
}

// WithHealth 设置 grpc 服务器是否开启健康检查。
func WithHealth(health bool) ServerOption {
	return func(options *serverOptions) {
		options.health = health
	}
}

// WithStreamServerInterceptors 使用给定的服务端 stream 拦截器。
func WithStreamServerInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.ServerOption {
	return grpc.ChainStreamInterceptor(interceptors...)
}

// WithUnaryServerInterceptors 使用给定的服务端 unary 拦截器。
func WithUnaryServerInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(interceptors...)
}
