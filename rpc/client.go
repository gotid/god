package rpc

import (
	"log"
	"time"

	"git.zc0901.com/go/god/rpc/internal/clientinterceptors"

	"git.zc0901.com/go/god/lib/discovery"
	"git.zc0901.com/go/god/rpc/internal"
	"git.zc0901.com/go/god/rpc/internal/auth"
	"google.golang.org/grpc"
)

var (
	// WithDialOption 自定义拨号选项。
	WithDialOption = internal.WithDialOption
	// WithNonBlock 设为非阻塞模式。
	WithNonBlock = internal.WithNonBlock
	// WithTimeout 自定义超时时间。
	WithTimeout = internal.WithTimeout
	// WithRetry 设为自动重连。
	WithRetry = internal.WithRetry
	// WithTransportCredentials 自定义安全拨号证书。
	WithTransportCredentials = internal.WithTransportCredentials
	// WithUnaryClientInterceptor 自定义自定义一元客户端拦截器。
	WithUnaryClientInterceptor = internal.WithUnaryClientInterceptor
)

type (
	// ClientOption 自定义客户端选项。
	ClientOption = internal.ClientOption

	// Client 表示一个RPC 客户端。
	Client = internal.Client

	// RpcClient 是一个RPC客户端。
	RpcClient struct {
		client Client
	}
)

func (rc *RpcClient) Conn() *grpc.ClientConn {
	return rc.client.Conn()
}

// MustNewClient 根据配置文件新建rpc客户端，出错直接退出。
func MustNewClient(conf ClientConf, options ...ClientOption) Client {
	cli, err := NewClient(conf, options...)
	if err != nil {
		log.Fatal(err)
	}

	return cli
}

// NewClient 返回一个新的 rpc 客户端。
func NewClient(c ClientConf, options ...ClientOption) (Client, error) {
	var opts []ClientOption
	if c.HasCredential() {
		opts = append(opts, WithDialOption(grpc.WithPerRPCCredentials(&auth.Credential{
			App:   c.App,
			Token: c.Token,
		})))
	}
	if c.NonBlock {
		opts = append(opts, WithNonBlock())
	}
	if c.Timeout > 0 {
		opts = append(opts, WithTimeout(time.Duration(c.Timeout)*time.Millisecond))
	}
	if c.Retry {
		opts = append(opts, WithRetry())
	}
	opts = append(opts, options...)

	var target string
	var err error
	if len(c.Endpoints) > 0 {
		target = internal.BuildDirectTarget(c.Endpoints)
	} else if len(c.Target) > 0 {
		target = c.Target
	} else {
		if err = c.Etcd.Validate(); err != nil {
			return nil, err
		}

		if c.Etcd.HasAccount() {
			discovery.RegisterAccount(c.Etcd.Hosts, c.Etcd.User, c.Etcd.Pass)
		}

		target = internal.BuildDiscoveryTarget(c.Etcd.Hosts, c.Etcd.Key)
	}
	client, err := internal.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}

	return &RpcClient{
		client: client,
	}, nil
}

// NewClientWithTarget 返回一个连接至 target 的 rpc 客户端。
func NewClientWithTarget(target string, opts ...ClientOption) (Client, error) {
	return internal.NewClient(target, opts...)
}

// NewClientNoAuth 新建无需鉴权的、基于etcd的rpc客户端
func NewClientNoAuth(c discovery.EtcdConf, opts ...ClientOption) (Client, error) {
	client, err := internal.NewClient(internal.BuildDiscoveryTarget(c.Hosts, c.Key), opts...)
	if err != nil {
		return nil, err
	}

	return &RpcClient{client: client}, nil
}

// SetClientSlowThreshold 设置慢调用时间阈值。
func SetClientSlowThreshold(duration time.Duration) {
	clientinterceptors.SetSlowThreshold(duration)
}
