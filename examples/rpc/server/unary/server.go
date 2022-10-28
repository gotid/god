package main

import (
	"context"
	"flag"
	"github.com/gotid/god/examples/rpc/remote/unary"
	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
	"os"
	"sync"
	"time"
)

var configFile = flag.String("f", "/Users/zs/Github/gotid/god/examples/rpc/server/unary/etc/config.yaml", "")

func main() {
	flag.Parse()

	var c rpc.ServerConfig
	conf.MustLoad(*configFile, &c)

	server := rpc.MustNewServer(c, func(grpcServer *grpc.Server) {
		//unary.RegisterGreeterServer(grpcServer, unary.UnimplementedGreeterServer{})
		unary.RegisterGreeterServer(grpcServer, NewGreetServer())
	})
	server.Start()
}

type GreetServer struct {
	unary.UnimplementedGreeterServer

	lock     sync.Mutex
	alive    bool
	downTime time.Time
}

func NewGreetServer() *GreetServer {
	return &GreetServer{
		alive: true,
	}
}

func (gs *GreetServer) Greet(ctx context.Context, request *unary.Request) (*unary.Response, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &unary.Response{
		Greet: "hi from " + hostname,
	}, nil
}
