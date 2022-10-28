package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gotid/god/examples/rpc/remote/unary"
	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/rpc"
	"time"
)

var configFile = flag.String("f", "/Users/zs/Github/gotid/god/examples/rpc/client/direct/config.yaml", "配置文件")

const timeFormat = "15:04:05"

func main() {
	flag.Parse()

	var c rpc.ClientConfig
	conf.MustLoad(*configFile, &c)

	client := rpc.MustNewClient(c)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn := client.Conn()
			greet := unary.NewGreeterClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			resp, err := greet.Greet(ctx, &unary.Request{Name: "richard"})
			if err != nil {
				fmt.Printf("%s X %s\n", time.Now().Format(timeFormat), err.Error())
			} else {
				fmt.Printf("%s => %s\n", time.Now().Format(timeFormat), resp.Greet)
			}
			cancel()
		}
	}
}
