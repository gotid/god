package main

import (
	"flag"
	"fmt"

	"github.com/gotid/god/examples/download/internal/config"
	"github.com/gotid/god/examples/download/internal/handler"
	"github.com/gotid/god/examples/download/internal/svc"

	"github.com/gotid/god/api"
	"github.com/gotid/god/lib/conf"
)

var configFile = flag.String("f", "etc/download-api.yaml", "配置文件")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := api.MustNewServer(c.Config)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("启动 api 服务器 %s:%d...\n", c.Host, c.Port)
	server.Start()
}
