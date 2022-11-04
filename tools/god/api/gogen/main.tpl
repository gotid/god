package main

import (
	"flag"
	"fmt"

	{{.importPackages}}
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "配置文件")

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
