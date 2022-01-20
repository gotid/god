package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/gotid/god/example/rpc/pb/stream"
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
	// client := rpc.MustNewClient(c)
	client, err := rpc.NewClientNoAuth(c.Etcd)
	if err != nil {
		log.Fatal(err)
	}

	// 新建流式rpc客户端
	conn := client.Conn()
	greeterClient := stream.NewStreamGreeterClient(conn)
	stm, err := greeterClient.Greet(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// rpc 流接收
	var wg sync.WaitGroup
	go func() {
		for {
			resp, err := stm.Recv()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("接收到=>", resp.Greet)
			wg.Done()
		}
	}()

	// rpc 流发送
	name := "richard"
	for i := 0; i < 3; i++ {
		wg.Add(1)
		fmt.Println("<=", name)
		if err := stm.Send(&stream.StreamReq{Name: name}); err != nil {
			log.Fatal(err)
		}
	}

	wg.Wait()
}
