package internal

import (
	"git.zc0901.com/go/god/lib/proc"
	"git.zc0901.com/go/god/lib/stat"
	"git.zc0901.com/go/god/rpc/internal/serverinterceptors"
	"google.golang.org/grpc"
	"net"
)

type (
	serverOptions struct {
		metrics *stat.Metrics
	}

	ServerOption func(options *serverOptions)

	server struct {
		*baseServer
		name string
	}
)

func init() {
	InitLogger()
}

func NewRpcServer(address string, opts ...ServerOption) Server {
	var options serverOptions
	for _, opt := range opts {
		opt(&options)
	}
	if options.metrics == nil {
		options.metrics = stat.NewMetrics(address)
	}

	return &server{
		baseServer: newBaseServer(address, options.metrics),
	}
}

// SetName 设置 Rpc 服务名称
func (s *server) SetName(name string) {
	s.name = name
	s.baseServer.SetName(name)
}

// Start 启动 Rpc 服务器监听
func (s *server) Start(register RegisterFn) error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	// 一元拦截器
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		serverinterceptors.UnaryTraceInterceptor(s.name),   // 链路跟踪
		serverinterceptors.UnaryCrashInterceptor(),         // 异常捕获
		serverinterceptors.UnaryStatInterceptor(s.metrics), // 数据统计
		serverinterceptors.UnaryPrometheusInterceptor(),    // 监控报警
	}
	unaryInterceptors = append(unaryInterceptors, s.unaryInterceptors...)

	// 流式拦截器
	streamInterceptors := []grpc.StreamServerInterceptor{
		serverinterceptors.StreamCrashInterceptor,
	}
	streamInterceptors = append(streamInterceptors, s.streamInterceptors...)

	// 设置自定义选项
	options := append(
		s.options,
		WithUnaryServerInterceptors(unaryInterceptors...),
		WithStreamServerInterceptors(streamInterceptors...),
	)
	server := grpc.NewServer(options...)
	register(server)

	// 平滑重启
	waitForCalled := proc.AddWrapUpListener(func() {
		server.GracefulStop()
	})
	defer waitForCalled()

	// 启动RPC服务监听
	return server.Serve(listener)
}

// WithMetrics 携带监控选项
func WithMetrics(metrics *stat.Metrics) ServerOption {
	return func(options *serverOptions) {
		options.metrics = metrics
	}
}
