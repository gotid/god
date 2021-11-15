package main

import (
	"context"
	"flag"

	"git.zc0901.com/go/god/example/rpc/pb/unary"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/service"
	"git.zc0901.com/go/god/rpc"
	"google.golang.org/grpc"
)

var (
	listen = flag.String("listen", "0.0.0.0:3456", "监听地址")
	server = flag.String("server", "dsn:///unaryserver:3456", "后端服务")
)

type GreetServer struct {
	*rpc.RpcProxy
}

func (s *GreetServer) Greet(ctx context.Context, req *unary.Request) (*unary.Response, error) {
	conn, err := s.TakeConn(ctx)
	if err != nil {
		return nil, err
	}

	remote := unary.NewGreeterClient(conn)
	return remote.Greet(ctx, req)
}

func main() {
	flag.Parse()

	// 新建代理客户端
	proxy := rpc.MustNewServer(rpc.ServerConf{
		ServiceConf: service.ServiceConf{
			Log: logx.LogConf{Mode: "console"},
		},
		ListenOn: *listen,
	}, func(grpcServer *grpc.Server) {
		unary.RegisterGreeterServer(grpcServer, &GreetServer{
			RpcProxy: rpc.NewProxy(*server),
		})
	})
	proxy.Start()
}
