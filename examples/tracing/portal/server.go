package main

import (
	"context"

	"github.com/gotid/god/examples/tracing/remote/portal"
	"github.com/gotid/god/examples/tracing/remote/user"
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
)

type PortalServer struct {
	portal.UnimplementedPortalServer

	userRpc rpc.Client
}

func (ps PortalServer) Portal(ctx context.Context, req *portal.PortalRequest) (*portal.PortalResponse, error) {
	conn := ps.userRpc.Conn()
	greet := user.NewUserClient(conn)
	resp, err := greet.GetGrade(ctx, &user.UserRequest{
		Name: req.Name,
	})
	if err != nil {
		return &portal.PortalResponse{
			Response: err.Error(),
		}, nil
	} else {
		return &portal.PortalResponse{
			Response: resp.Response,
		}, nil
	}
}

func NewPortalServer(client rpc.Client) *PortalServer {
	return &PortalServer{
		userRpc: client,
	}
}

func main() {
	client := rpc.MustNewClient(rpc.ClientConfig{
		Etcd: discov.EtcdConfig{
			Hosts: []string{"localhost:2379"},
			Key:   "user",
		},
	})
	server := rpc.MustNewServer(rpc.ServerConfig{
		Config: service.Config{
			Name: "portal.rpc",
		},
		ListenOn: "localhost:3456",
		Etcd: discov.EtcdConfig{
			Hosts: []string{"localhost:2379"},
			Key:   "portal",
		},
		Timeout: 500,
	}, func(grpcServer *grpc.Server) {
		portal.RegisterPortalServer(grpcServer, NewPortalServer(client))
	})
	server.Start()
}
