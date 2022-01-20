package main

import (
	"flag"
	"fmt"

	"git.zc0901.com/go/god/example/graceful/dns/api/internal/config"
	"git.zc0901.com/go/god/example/graceful/dns/api/internal/handler"
	"git.zc0901.com/go/god/example/graceful/dns/api/internal/svc"

	"git.zc0901.com/go/god/api"
	"git.zc0901.com/go/god/lib/conf"
)

var configFile = flag.String("f", "etc/graceful-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := api.MustNewServer(c.ServerConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
