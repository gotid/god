package redis

import (
	"crypto/tls"
	red "github.com/go-redis/redis/v8"
	"github.com/gotid/god/lib/syncx"
	"io"
)

const (
	defaultDatabase = 0
	maxRetries      = 3
	idleConns       = 8
)

var clientManager = syncx.NewResourceManager()

func getClient(r *Redis) (*red.Client, error) {
	val, err := clientManager.Get(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		client := red.NewClient(&red.Options{
			Addr:         r.Addr,
			Password:     r.Pass,
			DB:           defaultDatabase,
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

	return val.(*red.Client), nil
}
