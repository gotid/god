package rpc

import (
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/rpc/resolver"
)

type (
	// ServerConfig 是一个 RPC 服务端配置。
	ServerConfig struct {
		service.Config
		ListenOn      string
		Etcd          discov.EtcdConfig `json:",optional"`
		Auth          bool              `json:",optional"`
		Redis         redis.KeyConfig   `json:",optional"`
		StrictControl bool
		Timeout       int64 `json:",default=2000"`               // 连接超时阈值
		CpuThreshold  int64 `json:",default=900,range=[0,1000]"` // CPU泄流阈值
		Health        bool  `json:",default=true"`               // 服务是否健康
	}

	// ClientConfig 是一个 RPC 客户端配置。
	ClientConfig struct {
		Etcd      discov.EtcdConfig `json:",optional"`
		Endpoints []string          `json:",optional"` // 服务端地址
		Target    string            `json:",optional"`
		App       string            `json:",optional"`
		Token     string            `json:",optional"`
		NonBlock  bool              `json:",optional"` // 是否为非阻塞拨号
		Timeout   int64             `json:",default=2000"`
	}
)

// NewDirectClientConfig 返回一个直连 RPC 客户端配置。
func NewDirectClientConfig(endpoints []string, app, token string) ClientConfig {
	return ClientConfig{
		Endpoints: endpoints,
		App:       app,
		Token:     token,
	}
}

// NewEtcdClientConfig 返回一个通过 ETCD 连接的RPC客户端配置。
func NewEtcdClientConfig(hosts []string, key, app, token string) ClientConfig {
	return ClientConfig{
		Etcd: discov.EtcdConfig{
			Hosts: hosts,
			Key:   key,
		},
		App:   app,
		Token: token,
	}
}

// HasEtcd 判断服务端配置中是否有 ETCD 设置。
func (c ServerConfig) HasEtcd() bool {
	return len(c.Etcd.Hosts) > 0 && len(c.Etcd.Key) > 0
}

// Validate 判断服务端配置是否有效。
func (c ServerConfig) Validate() error {
	if !c.Auth {
		return nil
	}

	return c.Redis.Validate()
}

// BuildTarget 从给定的客户端配置构建 RPC 目标。
func (c ClientConfig) BuildTarget() (string, error) {
	if len(c.Endpoints) > 0 {
		return resolver.BuildDirectTarget(c.Endpoints), nil
	} else if len(c.Target) > 0 {
		return c.Target, nil
	}

	if err := c.Etcd.Validate(); err != nil {
		return "", err
	}

	if c.Etcd.HasAccount() {
		discov.RegisterAccount(c.Etcd.Hosts, c.Etcd.User, c.Etcd.Pass)
	}

	if c.Etcd.HasTLS() {
		if err := discov.RegisterTLS(c.Etcd.Hosts, c.Etcd.CertFile, c.Etcd.CertKeyFile,
			c.Etcd.CACertFile, c.Etcd.InsecureSkipVerify); err != nil {
			return "", err
		}
	}

	return resolver.BuildDiscovTarget(c.Etcd.Hosts, c.Etcd.Key), nil
}

// HasCredential 检测配置中是否有证书设置。
func (c ClientConfig) HasCredential() bool {
	return len(c.App) > 0 && len(c.Token) > 0
}
