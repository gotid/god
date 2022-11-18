package main

import (
	"flag"
	"fmt"

	"github.com/gotid/god/examples/shorturl/rpc/transformer/internal/config"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/internal/server"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/internal/svc"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"

	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/transformer.yaml", "配置文件")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := rpc.MustNewServer(c.ServerConfig, func(grpcServer *grpc.Server) {
		transformer.RegisterTransformerServer(grpcServer, server.NewTransformerServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("启动 rpc 服务器 %s...\n", c.ListenOn)
	s.Start()
}
