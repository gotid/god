package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gotid/god/example/rpc/pb/unary"
	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/config.yaml", "配置文件")

type GreetServer struct{}

func (gs *GreetServer) Greet(ctx context.Context, req *unary.Request) (*unary.Response, error) {
	fmt.Println("=>", req)

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &unary.Response{Greet: hostname + "发来了问候"}, nil
}

func NewGreetServer() *GreetServer {
	return &GreetServer{}
}

func main() {
	// 加载配置
	flag.Parse()
	var c rpc.ServerConf
	conf.MustLoad(*configFile, &c)

	// 新建rpc服务器
	server := rpc.MustNewServer(c, func(server *grpc.Server) {
		unary.RegisterGreeterServer(server, NewGreetServer())
	})

	// 添加自定义拦截器
	server.AddUnaryInterceptors(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()
		resp, err = handler(ctx, req)
		log.Printf("方法：%s 耗时：%v\n", info.FullMethod, time.Since(startTime))
		return resp, err
	})
	server.Start()
}
