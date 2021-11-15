package internal

import (
	"git.zc0901.com/go/god/lib/stat"
	"google.golang.org/grpc"
)

type (
	RegisterFn func(server *grpc.Server)

	Server interface {
		AddOptions(options ...grpc.ServerOption)
		AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor)
		AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor)
		SetName(string)
		Start(register RegisterFn) error
	}

	baseServer struct {
		address            string
		metrics            *stat.Metrics
		maxRetries         int
		options            []grpc.ServerOption
		unaryInterceptors  []grpc.UnaryServerInterceptor
		streamInterceptors []grpc.StreamServerInterceptor
	}
)

func newBaseServer(address string, opts *serverOptions) *baseServer {
	return &baseServer{
		address:    address,
		metrics:    opts.metrics,
		maxRetries: opts.maxRetries,
	}
}

func (bs *baseServer) AddOptions(options ...grpc.ServerOption) {
	bs.options = append(bs.options, options...)
}

func (bs *baseServer) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	bs.unaryInterceptors = append(bs.unaryInterceptors, interceptors...)
}

func (bs *baseServer) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	bs.streamInterceptors = append(bs.streamInterceptors, interceptors...)
}

func (bs *baseServer) SetName(name string) {
	bs.metrics.SetName(name)
}
