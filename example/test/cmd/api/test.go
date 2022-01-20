package main

import (
	"flag"
	"fmt"

	"git.zc0901.com/go/god/example/test/cmd/api/internal/config"
	"git.zc0901.com/go/god/example/test/cmd/api/internal/handler"
	"git.zc0901.com/go/god/example/test/cmd/api/internal/svc"

	"git.zc0901.com/go/god/api"
	"git.zc0901.com/go/god/lib/conf"
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
