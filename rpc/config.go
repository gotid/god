package rpc

import (
	"github.com/gotid/god/lib/discovery"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/lib/store/redis"
)

type (
	// ServerConf Rpc服务端配置
	ServerConf struct {
		service.ServiceConf                    // 服务配置
		ListenOn            string             // rpc监听地址和端口，如：127.0.0.1:8888
		Etcd                discovery.EtcdConf `json:",optional"`                   // etcd相关配置
		Auth                bool               `json:",optional"`                   // 是否开启rpc通信鉴权，若是则Redis配置必填
		Redis               redis.KeyConf      `json:",optional"`                   // rpc通信及安全的redis配置
		StrictControl       bool               `json:",optional"`                   // 是否为严格模式，若是且遇到鉴权错误则抛出异常
		Timeout             int64              `json:",default=2000"`               // 默认超时时长为2000毫秒，<=0则表示无超时限制
		CpuThreshold        int64              `json:",default=900,range=[0:1000]"` // cpu降载阈值，默认900，支持区间为0-1000
	}

	// ClientConf Rpc客户端配置
	ClientConf struct {
		Etcd      discovery.EtcdConf `json:",optional"`
		Endpoints []string           `json:",optional"`
		Target    string             `json:",optional"`
		App       string             `json:",optional"`
		Token     string             `json:",optional"`
		NonBlock  bool               `json:",optional"`
		Timeout   int64              `json:",default=2000"` // 连接超时，单位毫秒
	}
)

// NewDirectClientConf 新建Rpc直联客户端配置
func NewDirectClientConf(endpoints []string, app, token string) ClientConf {
	return ClientConf{
		Endpoints: endpoints,
		App:       app,
		Token:     token,
	}
}

// NewEtcdClientConf 新建Rpc Etcd连接客户端配置
func NewEtcdClientConf(hosts []string, key, app, token string) ClientConf {
	return ClientConf{
		Etcd: discovery.EtcdConf{
			Hosts: hosts,
			Key:   key,
		},
		App:   app,
		Token: token,
	}
}

// HasEtcd 判断服务端配置是否有Etcd主机和键。
func (sc ServerConf) HasEtcd() bool {
	return len(sc.Etcd.Hosts) > 0 && len(sc.Etcd.Key) > 0
}

// Validate 验证鉴权情况下redis的主机和部署模式是否已提供
func (sc ServerConf) Validate() error {
	if !sc.Auth {
		return nil
	}

	return sc.Redis.Validate()
}

// HasCredential 判断客户端配置是否存在App+token凭证
func (cc ClientConf) HasCredential() bool {
	return len(cc.App) > 0 && len(cc.Token) > 0
}
