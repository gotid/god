package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/credentials"

	"git.zc0901.com/go/god/rpc/internal/balancer/p2c"
	"git.zc0901.com/go/god/rpc/internal/clientinterceptors"
	"git.zc0901.com/go/god/rpc/internal/resolver"
	"google.golang.org/grpc"
)

const (
	dialTimeout = 3 * time.Second
	separator   = '/'
)

type (
	// Client 包装 Conn 方法的接口
	Client interface {
		Conn() *grpc.ClientConn
	}

	// ClientOptions 是RPC客户端选择项
	ClientOptions struct {
		NonBlock    bool
		Secure      bool
		Retry       bool
		Timeout     time.Duration
		DialOptions []grpc.DialOption
	}

	// ClientOption 是自定义客户端选项的方法
	ClientOption func(options *ClientOptions)

	client struct {
		conn *grpc.ClientConn
	}
)

func init() {
	// 注册服务直连和服务发现两种模式
	resolver.RegisterResolver()
}

// NewClient 返回RPC客户端
func NewClient(target string, opts ...ClientOption) (Client, error) {
	var cli client
	opts = append([]ClientOption{WithDialOption(grpc.WithBalancerName(p2c.Name))}, opts...)
	if err := cli.dial(target, opts...); err != nil {
		return nil, err
	}

	return &cli, nil
}

func (c *client) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *client) buildDialOptions(opts ...ClientOption) []grpc.DialOption {
	var cliOpts ClientOptions
	for _, opt := range opts {
		opt(&cliOpts)
	}

	var options []grpc.DialOption
	if !cliOpts.Secure {
		options = append([]grpc.DialOption(nil), grpc.WithInsecure())
	}

	if !cliOpts.NonBlock {
		options = append(options, grpc.WithBlock())
	}

	options = append(options,
		WithUnaryClientInterceptors(
			clientinterceptors.UnaryTraceInterceptor,               // 线路跟踪
			clientinterceptors.DurationInterceptor,                 // 慢查询日志
			clientinterceptors.PrometheusInterceptor,               // 监控报警
			clientinterceptors.BreakerInterceptor,                  // 自动熔断
			clientinterceptors.TimeoutInterceptor(cliOpts.Timeout), // 超时控制
			clientinterceptors.RetryInterceptor(cliOpts.Retry),     // 连接重试
		),
		WithStreamClientInterceptors(
			clientinterceptors.StreamTraceInterceptor,
		),
	)

	return append(options, cliOpts.DialOptions...)
}

func (c *client) dial(server string, opts ...ClientOption) error {
	options := c.buildDialOptions(opts...)
	timeCtx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(timeCtx, server, options...)
	if err != nil {
		service := server
		if errors.Is(err, context.DeadlineExceeded) {
			pos := strings.LastIndexByte(server, separator)
			if 0 < pos && pos < len(server)-1 {
				service = server[pos+1:]
			}
		}
		return fmt.Errorf("RPC 拨号失败: %s, 错误：%s, 请确保 RPC %s 已启动，如使用 etcd 也需确保启动",
			server, err.Error(), service)
	}

	c.conn = conn
	return nil
}

// WithDialOption 自定义 grpc.DialOption。
func WithDialOption(opt grpc.DialOption) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, opt)
	}
}

// WithNonBlock 设为非阻塞模式。
func WithNonBlock() ClientOption {
	return func(options *ClientOptions) {
		options.NonBlock = true
	}
}

// WithTimeout 设置超时时间。
func WithTimeout(timeout time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.Timeout = timeout
	}
}

// WithRetry 设为自动重连。
func WithRetry() ClientOption {
	return func(options *ClientOptions) {
		options.Retry = true
	}
}

// WithUnaryClientInterceptor 自定义一元客户端拦截器。
func WithUnaryClientInterceptor(interceptor grpc.UnaryClientInterceptor) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, WithUnaryClientInterceptors(interceptor))
	}
}

// WithTransportCredentials 自定义安全拨号证书。
func WithTransportCredentials(creds credentials.TransportCredentials) ClientOption {
	return func(options *ClientOptions) {
		options.Secure = true
		options.DialOptions = append(options.DialOptions, grpc.WithTransportCredentials(creds))
	}
}
