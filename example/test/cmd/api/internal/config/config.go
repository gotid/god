package config

import (
	"github.com/gotid/god/api"
	"github.com/gotid/god/rpc"
)

type Config struct {
	api.ServerConf
	TestRPC rpc.ClientConf
}
