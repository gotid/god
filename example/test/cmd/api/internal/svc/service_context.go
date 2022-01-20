package svc

import (
	"github.com/gotid/god/example/test/cmd/api/internal/config"
	"github.com/gotid/god/example/test/cmd/rpc/testclient"
	"github.com/gotid/god/rpc"
)

type ServiceContext struct {
	Config  config.Config
	TestRPC testclient.Test
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		TestRPC: testclient.NewTest(rpc.MustNewClient(c.TestRPC)),
	}
}
