package main

import (
	"context"
	"flag"
	"github.com/gotid/god/examples/rpc/remote/unary"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
)

var (
	listen = flag.String("listen", "0.0.0.0:3457", "监听地址")
	server = flag.String("server", "dns:///unaryserver:3456", "后台服务")
)

type GreetServer struct {
	*unary.UnimplementedGreeterServer
	*rpc.Proxy
}

func (s *GreetServer) Greet(ctx context.Context, request *unary.Request) (*unary.Response, error) {
	conn, err := s.TakeConn(ctx)
	if err != nil {
		return nil, err
	}

	client := unary.NewGreeterClient(conn)
	return client.Greet(ctx, request)
}

func main() {
	flag.Parse()

	proxy := rpc.MustNewServer(rpc.ServerConfig{
		Config: service.Config{
			Log: logx.Config{
				Mode: "console",
			},
		},
		ListenOn: *listen,
	}, func(grpcServer *grpc.Server) {
		unary.RegisterGreeterServer(grpcServer, &GreetServer{
			Proxy: rpc.NewProxy(*server),
		})
	})
	proxy.Start()
}
