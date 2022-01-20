package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gotid/god/example/rpc/pb/unary"
	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/rpc"
)

var configFile = flag.String("f", "config.yaml", "配置文件")

func main() {
	// 加载配置
	flag.Parse()
	var c rpc.ClientConf
	conf.MustLoad(*configFile, &c)

	// 新建rpc客户端
	client := rpc.MustNewClient(c)
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	// 运行
	for {
		select {
		case <-ticker.C:
			conn := client.Conn()
			greeterClient := unary.NewGreeterClient(conn)
			resp, err := greeterClient.Greet(context.Background(), &unary.Request{
				Name: "kevin",
			})
			if err != nil {
				fmt.Println("错误：", err.Error())
			} else {
				fmt.Println("=>", resp.Greet)
			}
		}
	}
}
