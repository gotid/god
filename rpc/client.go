package rpc

import (
	"github.com/gotid/god/rpc/internal"
	"github.com/gotid/god/rpc/internal/auth"
	"github.com/gotid/god/rpc/internal/clientinterceptors"
	"google.golang.org/grpc"
	"log"
	"time"
)

var (
	// WithDialOption 是 internal.WithDialOption 的别名。
	WithDialOption = internal.WithDialOption
	// WithNonBlock 将拨号设置为非阻塞模式。
	WithNonBlock = internal.WithNonBlock
	// WithStreamClientInterceptor 是 internal.WithStreamClientInterceptor 的别名。
	WithStreamClientInterceptor = internal.WithStreamClientInterceptor
	// WithTimeout 是 internal.WithTimeout 的别名。
	WithTimeout = internal.WithTimeout
	// WithTransportCredentials 确保 gRPC 调用使用给定的凭据进行加密。
	WithTransportCredentials = internal.WithTransportCredentials
	// WithUnaryClientInterceptor 是 internal.WithUnaryClientInterceptor 的别名。
	WithUnaryClientInterceptor = internal.WithUnaryClientInterceptor
)

type (
	// Client 是 internal.Client 的别名。
	Client = internal.Client
	// ClientOption 是 internal.ClientOption 的别名。
	ClientOption = internal.ClientOption

	// RpcClient 是一个 rpc 客户端。
	RpcClient struct {
		client Client
	}
)

// Conn 返回底层 grpc.ClientConn 连接。
func (c *RpcClient) Conn() *grpc.ClientConn {
	return c.client.Conn()
}

// MustNewClient 返回一个 rpc 客户端 Client，遇错退出。
func MustNewClient(config ClientConfig, options ...ClientOption) Client {
	cli, err := NewClient(config, options...)
	if err != nil {
		log.Fatal(err)
	}

	return cli
}

// NewClient 返回一个 rpc 客户端 Client。
func NewClient(c ClientConfig, options ...ClientOption) (Client, error) {
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
	opts = append(opts, options...)

	target, err := c.BuildTarget()
	if err != nil {
		return nil, err
	}

	client, err := internal.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}

	return &RpcClient{client: client}, nil
}

// NewClientWithTarget 返回给定目标的 rpc 客户端 Client。
func NewClientWithTarget(target string, opts ...ClientOption) (Client, error) {
	return internal.NewClient(target, opts...)
}

// SetClientSlowThreshold 设置客户端慢调用时长。
func SetClientSlowThreshold(threshold time.Duration) {
	clientinterceptors.SetSlowThreshold(threshold)
}

// DontLogContentMethod 不记录给定方法的请求/响应详情。
func DontLogContentMethod(method string) {
	clientinterceptors.DontLogContentMethod(method)
}
