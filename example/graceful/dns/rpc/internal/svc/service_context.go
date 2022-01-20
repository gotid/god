package svc

import "github.com/gotid/god/example/graceful/dns/rpc/internal/config"

type ServiceContext struct {
	c config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		c: c,
	}
}
