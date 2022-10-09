package redis

import (
	"crypto/tls"
	red "github.com/go-redis/redis/v8"
	"github.com/gotid/god/lib/syncx"
	"io"
)

var clusterManager = syncx.NewResourceManager()

func getCluster(r *Redis) (*red.ClusterClient, error) {
	val, err := clusterManager.Get(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		client := red.NewClusterClient(&red.ClusterOptions{
			Addrs:        []string{r.Addr},
			Password:     r.Pass,
			MaxRetries:   maxRetries,
			MinIdleConns: idleConns,
			TLSConfig:    tlsConfig,
		})
		client.AddHook(durationHook)

		return client, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*red.ClusterClient), nil
}
