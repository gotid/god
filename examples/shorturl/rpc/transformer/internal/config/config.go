package config

import (
	"github.com/gotid/god/lib/store/cache"
	"github.com/gotid/god/rpc"
)

type Config struct {
	rpc.ServerConfig

	DataSource string
	Table      string
	Cache      cache.Config
}
