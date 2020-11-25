package redis

import (
	"fmt"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/mapping"
	"git.zc0901.com/go/god/lib/syncx"
	"git.zc0901.com/go/god/lib/timex"
	"github.com/go-redis/redis"
	"io"
	"strings"
)

var (
	clusterClientManager    = syncx.NewResourceManager()
	standaloneClientManager = syncx.NewResourceManager()
)

func getClient(r *Redis) (Client, error) {
	switch r.Mode {
	case ClusterMode:
		return getClusterClient(r.Addr, r.Password)
	case StandaloneMode:
		return getStandaloneClient(r.Addr, r.Password)
	default:
		return nil, fmt.Errorf("不支持的 redis 模式 '%s'", r.Mode)
	}
}

func getClusterClient(addr, password string) (Client, error) {
	client, err := clusterClientManager.Get(addr, func() (io.Closer, error) {
		client := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        []string{addr},
			Password:     password,
			MaxRetries:   maxRetries,
			MinIdleConns: idleConns,
		})
		client.WrapProcess(process)
		return client, nil
	})
	if err != nil {
		return nil, err
	}

	return client.(*redis.ClusterClient), nil
}

func getStandaloneClient(addr, password string) (Client, error) {
	client, err := standaloneClientManager.Get(addr, func() (io.Closer, error) {
		client := redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     password,
			DB:           defaultDatabase,
			MaxRetries:   maxRetries,
			MinIdleConns: idleConns,
		})
		client.WrapProcess(process)
		return client, nil
	})
	if err != nil {
		return nil, err
	}

	return client.(*redis.Client), nil
}

// 包装redis执行命令，采集慢查询日志
func process(proc func(redis.Cmder) error) func(redis.Cmder) error {
	return func(cmd redis.Cmder) error {
		start := timex.Now()

		defer func() {
			duration := timex.Since(start)
			if duration > slowThreshold {
				var b strings.Builder
				for i, arg := range cmd.Args() {
					if i > 0 {
						b.WriteByte(' ')
					}
					b.WriteString(mapping.Repr(arg))
				}
				logx.WithDuration(duration).Slowf("[REDIS] 慢查询 - %s", b.String())
			}
		}()

		return proc(cmd)
	}
}
