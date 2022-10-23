package internal

import (
	"github.com/gotid/god/lib/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/keepalive"
	"time"
)

const defaultConnectionIdleDuration = 5 * time.Minute

type (
	// RegisterFn 定义注册一个 rpc 服务器的方法。
	RegisterFn func(server *grpc.Server)

	// Server 接口表示一个 rpc 服务器。
	Server interface {
		// AddOptions 添加 rpc 服务器选项。
		AddOptions(options ...grpc.ServerOption)
		// AddStreamInterceptors 添加流式拦截器。
		AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor)
		// AddUnaryInterceptors 添加意愿拦截器。
		AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor)
		// SetName 设置 rpc 名称。
		SetName(name string)
		// Start 用给定的注册函数启动 rpc 服务器。
		Start(register RegisterFn) error
	}

	baseServer struct {
		address            string
		health             *health.Server
		metrics            *stat.Metrics
		options            []grpc.ServerOption
		streamInterceptors []grpc.StreamServerInterceptor
		unaryInterceptors  []grpc.UnaryServerInterceptor
	}
)

func newBaseServer(address string, options *serverOptions) *baseServer {
	var h *health.Server
	if options.health {
		h = health.NewServer()
	}
	return &baseServer{
		address: address,
		health:  h,
		metrics: options.metrics,
		options: []grpc.ServerOption{grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: defaultConnectionIdleDuration,
		})},
	}
}

func (s *baseServer) AddOptions(options ...grpc.ServerOption) {
	s.options = append(s.options, options...)
}

func (s *baseServer) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	s.streamInterceptors = append(s.streamInterceptors, interceptors...)
}

func (s *baseServer) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
}

func (s *baseServer) SetName(name string) {
	s.metrics.SetName(name)
}
