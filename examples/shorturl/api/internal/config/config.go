package config

import (
	"github.com/gotid/god/api"
	"github.com/gotid/god/rpc"
)

type Config struct {
	api.Config

	Transformer rpc.ClientConfig
}
