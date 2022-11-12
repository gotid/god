package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/examples/gateway/hello"
	"github.com/gotid/god/examples/gateway/internal/config"
	"github.com/gotid/god/examples/gateway/internal/server"
	"github.com/gotid/god/examples/gateway/internal/svc"

	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "/Users/zs/Github/gotid/god/examples/gateway/etc/hello.yaml", "配置文件")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := rpc.MustNewServer(c.ServerConfig, func(grpcServer *grpc.Server) {
		hello.RegisterHelloServer(grpcServer, server.NewHelloServer(ctx))

		//if c.Mode == service.DevMode || c.Mode == service.TestMode {
		reflection.Register(grpcServer)
		//}
	})
	defer s.Stop()

	fmt.Printf("启动 rpc 服务器 %s...\n", c.ListenOn)
	s.Start()
}
