package config

import (
	"git.zc0901.com/go/god/api"
	"git.zc0901.com/go/god/rpc"
)

type Config struct {
	api.ServerConf
	TestRPC rpc.ClientConf
}
