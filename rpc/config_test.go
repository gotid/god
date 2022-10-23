package rpc

import (
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/lib/store/redis"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClientConfig(t *testing.T) {
	config := NewDirectClientConfig([]string{"localhost:1314"}, "foo", "bar")
	assert.True(t, config.HasCredential())
	config = NewEtcdClientConfig([]string{"localhost:1314", "localhost:1213"}, "key", "foo", "bar")
	assert.True(t, config.HasCredential())
}

func TestServerConfig(t *testing.T) {
	config := ServerConfig{
		Config:   service.Config{},
		ListenOn: "",
		Etcd: discov.EtcdConfig{
			Hosts: []string{"localhost:1234"},
			Key:   "key",
		},
		Auth: true,
		Redis: redis.KeyConfig{
			Config: redis.Config{
				Type: redis.NodeType,
			},
			Key: "foo",
		},
		StrictControl: false,
		Timeout:       0,
		CpuThreshold:  0,
	}
	assert.True(t, config.HasEtcd())
	assert.NotNil(t, config.Validate())
	config.Redis.Host = "localhost:5678"
	assert.Nil(t, config.Validate())
}
