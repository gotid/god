package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gotid/god/examples/rpc/remote/unary"
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/rpc"
	"log"
	"time"
)

var lb = flag.String("t", "direct", "负载均衡类型")

func main() {
	flag.Parse()

	var cli rpc.Client
	switch *lb {
	case "direct":
		cli = rpc.MustNewClient(rpc.ClientConfig{
			Endpoints: []string{
				"localhost:3456",
				"localhost:3457",
			},
		})
	case "discov":
		cli = rpc.MustNewClient(rpc.ClientConfig{
			Etcd: discov.EtcdConfig{
				Hosts: []string{"localhost:2379"},
				Key:   "rpc.unary",
			},
		})
	default:
		log.Fatal("错误的负载均衡类型")
	}

	greet := unary.NewGreeterClient(cli.Conn())
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resp, err := greet.Greet(context.Background(), &unary.Request{
				Name: "richard",
			})
			if err != nil {
				fmt.Println("X", err.Error())
			} else {
				fmt.Println("=>", resp.Greet)
			}
		}
	}
}
