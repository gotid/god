package main

import (
	"fmt"
	"io"

	"github.com/gotid/god/example/rpc/pb/stream"
	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
)

type StreamGreetServer int

func (s StreamGreetServer) Greet(server stream.StreamGreeter_GreetServer) error {
	ctx := server.Context()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("客户端取消")
			return ctx.Err()
		default:
			req, err := server.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Println("=>", req.Name)
			greet := "hello, " + req.Name
			fmt.Println("<=", greet)
			server.Send(&stream.StreamResp{Greet: greet})
		}
	}
}

func main() {
	var c rpc.ServerConf
	conf.MustLoad("etc/config.yaml", &c)
	server := rpc.MustNewServer(c, func(server *grpc.Server) {
		stream.RegisterStreamGreeterServer(server, StreamGreetServer(0))
	})
	server.Start()
}
