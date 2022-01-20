package redis

import (
	"crypto/tls"
	"fmt"
	"io"
	"strings"

	"github.com/go-redis/redis"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mapping"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
)

const (
	defaultDatabase = 0
	maxRetries      = 3
	idleConns       = 8
)

var (
	clusterClientManager    = syncx.NewResourceManager()
	standaloneClientManager = syncx.NewResourceManager()
)

func getClient(r *Redis) (Client, error) {
	switch r.Mode {
	case ClusterMode:
		return getClusterClient(r)
	case StandaloneMode:
		return getStandaloneClient(r)
	default:
		return nil, fmt.Errorf("不支持的 redis 模式 '%s'", r.Mode)
	}
}

func getClusterClient(r *Redis) (Client, error) {
	client, err := clusterClientManager.Get(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		client := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        []string{r.Addr},
			Password:     r.Password,
			MaxRetries:   maxRetries,
			MinIdleConns: idleConns,
			TLSConfig:    tlsConfig,
		})
		client.WrapProcess(process)
		return client, nil
	})
	if err != nil {
		return nil, err
	}

	return client.(*redis.ClusterClient), nil
}

func getStandaloneClient(r *Redis) (Client, error) {
	client, err := standaloneClientManager.Get(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		client := redis.NewClient(&redis.Options{
			Addr:         r.Addr,
			Password:     r.Password,
			DB:           defaultDatabase,
			MaxRetries:   maxRetries,
			MinIdleConns: idleConns,
			TLSConfig:    tlsConfig,
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
