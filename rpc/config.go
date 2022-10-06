package rpc

import (
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/lib/store/redis"
)

type (
	// ServerConfig 是一个 rpc 服务端配置。
	ServerConfig struct {
		service.Config
		ListenOn string
		Etcd     discov.EtcdConfig `json:",optional"`
		Auth     bool              `json:",optional"`
		Redis    redis.KeyConfig   `json:",optional"`
	}
)
