package svc

import (
	"github.com/gotid/god/examples/shorturl/api/internal/config"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformerclient"
	"github.com/gotid/god/rpc"
)

type ServiceContext struct {
	Config      config.Config
	Transformer transformerclient.Transformer
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		Transformer: transformerclient.NewTransformer(rpc.MustNewClient(c.Transformer)),
	}
}
