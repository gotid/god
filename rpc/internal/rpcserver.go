package internal

import (
	"net"

	"git.zc0901.com/go/god/lib/proc"
	"git.zc0901.com/go/god/lib/stat"
	"git.zc0901.com/go/god/rpc/internal/serverinterceptors"
	"google.golang.org/grpc"
)

type (
	serverOptions struct {
		metrics    *stat.Metrics
		maxRetries int
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

// NewRpcServer 返回一个新的 RPC 服务器。
func NewRpcServer(address string, opts ...ServerOption) Server {
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
		serverinterceptors.UnaryTraceInterceptor,           // 链路跟踪
		serverinterceptors.RetryInterceptor(s.maxRetries),  // 连接重试
		serverinterceptors.UnaryCrashInterceptor,           // 异常捕获
		serverinterceptors.UnaryStatInterceptor(s.metrics), // 数据统计
		serverinterceptors.UnaryPrometheusInterceptor,      // 监控报警
		serverinterceptors.UnaryBreakerInterceptor,
	}
	unaryInterceptors = append(unaryInterceptors, s.unaryInterceptors...)

	// 流式拦截器
	streamInterceptors := []grpc.StreamServerInterceptor{
		serverinterceptors.StreamTracingInterceptor,
		serverinterceptors.StreamCrashInterceptor,
		serverinterceptors.StreamBreakerInterceptor,
	}
	streamInterceptors = append(streamInterceptors, s.streamInterceptors...)

	// 设置自定义选项
	options := append(
		s.options,
		WithUnaryServerInterceptors(unaryInterceptors...),
		WithStreamServerInterceptors(streamInterceptors...),
	)
	srv := grpc.NewServer(options...)
	register(srv)

	// 平滑重启
	waitForCalled := proc.AddWrapUpListener(func() {
		srv.GracefulStop()
	})
	defer waitForCalled()

	// 启动RPC服务监听
	return srv.Serve(listener)
}

// WithMetrics 携带监控选项
func WithMetrics(metrics *stat.Metrics) ServerOption {
	return func(options *serverOptions) {
		options.metrics = metrics
	}
}

// WithMaxRetries 自定义连接重试次数。
func WithMaxRetries(maxRetries int) ServerOption {
	return func(options *serverOptions) {
		options.maxRetries = maxRetries
	}
}
