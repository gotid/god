package cache

import "github.com/gotid/god/lib/store/redis"

type (
	// Config 是缓存集群节点配置 ClusterConfig 别名。
	Config = ClusterConfig

	// ClusterConfig 缓存集群节点配置。
	ClusterConfig []NodeConfig

	// NodeConfig 缓存节点配置。
	NodeConfig struct {
		redis.Config
		Weight int `json:",default=100"`
	}
)
