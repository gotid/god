package main

import (
	"flag"
	"fmt"

	"github.com/gotid/god/example/test/cmd/api/internal/config"
	"github.com/gotid/god/example/test/cmd/api/internal/handler"
	"github.com/gotid/god/example/test/cmd/api/internal/svc"

	"github.com/gotid/god/api"
	"github.com/gotid/god/lib/conf"
)

var configFile = flag.String("f", "etc/test.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := api.MustNewServer(c.ServerConf)
	defer server.Stop()

	// 使用上下文转metadata中间件
	// server.Use(http.ContextToMetadata)

	//// 设置错误处理函数
	//httpx.SetErrorHandler(Fail)
	//
	//// 设置成功处理函数
	//httpx.SetOkJsonHandler(Success)

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
