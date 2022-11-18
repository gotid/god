package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gotid/god/examples/tracing/remote/user"
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
)

type UserServer struct {
	user.UnimplementedUserServer

	lock     sync.Mutex
	alive    bool
	downTime time.Time
}

func (us *UserServer) GetGrade(ctx context.Context, req *user.UserRequest) (*user.UserResponse, error) {
	fmt.Println("=>", req)

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &user.UserResponse{
		Response: "hello from " + hostname,
	}, nil
}

func NewUserServer() *UserServer {
	return &UserServer{
		alive: true,
	}
}

func main() {
	server := rpc.MustNewServer(rpc.ServerConfig{
		Config: service.Config{
			Name: "user.rpc",
		},
		ListenOn: "localhost:3457",
		Etcd: discov.EtcdConfig{
			Hosts: []string{"localhost:2379"},
			Key:   "user",
		},
		Timeout: 500,
	}, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, NewUserServer())
	})
	defer server.Stop()
	server.Start()
}
