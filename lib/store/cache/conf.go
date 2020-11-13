package cache

import "git.zc0901.com/go/god/lib/store/redis"

type (
	// 节点配置
	Conf struct {
		redis.Conf
		Weight int `json:",default=100"`
	}

	// 集群配置
	ClusterConf []Conf
)
