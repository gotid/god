package redis

import (
	"fmt"
	red "github.com/go-redis/redis/v8"
	"github.com/gotid/god/lib/logx"
)

// ClosableNode 接口表示一个可关闭的节点。
type ClosableNode interface {
	Node
	Close()
}

// CreateBlockingNode 返回一个可关闭的阻塞节点 ClosableNode。
func CreateBlockingNode(r *Redis) (ClosableNode, error) {
	timeout := readWriteTimeout + blockingQueryTimeout

	switch r.Type {
	case NodeType:
		client := red.NewClient(&red.Options{
			Addr:         r.Addr,
			Password:     r.Pass,
			DB:           defaultDatabase,
			MaxRetries:   maxRetries,
			PoolSize:     1,
			MinIdleConns: 1,
			ReadTimeout:  timeout,
		})
		return &clientBridge{client}, nil
	case ClusterType:
		client := red.NewClusterClient(&red.ClusterOptions{
			Addrs:        []string{r.Addr},
			Password:     r.Pass,
			MaxRetries:   maxRetries,
			PoolSize:     1,
			MinIdleConns: 1,
			ReadTimeout:  timeout,
		})
		return &clusterBridge{client}, nil
	default:
		return nil, fmt.Errorf("未知的 redis 类型: %s", r.Type)
	}
}

type clientBridge struct {
	*red.Client
}

func (b *clientBridge) Close() {
	if err := b.Client.Close(); err != nil {
		logx.Errorf("关闭 redis 客户端时出错：%s", err)
	}
}

type clusterBridge struct {
	*red.ClusterClient
}

func (b *clusterBridge) Close() {
	if err := b.ClusterClient.Close(); err != nil {
		logx.Errorf("关闭 redis 集群客户端时出错：%s", err)
	}
}
