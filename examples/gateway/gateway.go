package main

import (
	"flag"
	"github.com/gotid/god/gateway"
	"github.com/gotid/god/lib/conf"
)

var gwConfigFile = flag.String("f", "/Users/zs/Github/gotid/god/examples/gateway/etc/gateway.yaml", "配置文件")

func main() {
	flag.Parse()

	var c gateway.Config
	conf.MustLoad(*gwConfigFile, &c)
	gw := gateway.MustNewServer(c)
	defer gw.Stop()
	gw.Start()
}
