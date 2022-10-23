package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotid/god/rpc/internal/balancer/p2c"
	"github.com/gotid/god/rpc/internal/clientinterceptors"
	"github.com/gotid/god/rpc/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"strings"
	"time"
)

const (
	dialTimeout = 3 * time.Second
	separator   = '/'
)

func init() {
	resolver.Register()
}

type (
	// Client 接口包装 Conn 方法。
	Client interface {
		Conn() *grpc.ClientConn
	}

	// ClientOptions 是一个客户端可选项。
	ClientOptions struct {
		NonBlock    bool
		Timeout     time.Duration
		Secure      bool
		DialOptions []grpc.DialOption
	}

	// ClientOption 自定义 ClientOptions 的方法。
	ClientOption func(options *ClientOptions)

	client struct {
		conn *grpc.ClientConn
	}
)

// NewClient 返回一个 Client。
func NewClient(target string, opts ...ClientOption) (Client, error) {
	var cli client

	svcCfg := fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, p2c.Name)
	balancerOpt := WithDialOption(grpc.WithDefaultServiceConfig(svcCfg))
	opts = append([]ClientOption{balancerOpt}, opts...)
	if err := cli.dial(target, opts...); err != nil {
		return nil, err
	}
	return &cli, nil
}

func (c *client) Conn() *grpc.ClientConn {
	return c.conn
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
			// len(server) - 1 是最后一个字符串的索引
			if 0 < pos && pos < len(server)-1 {
				service = server[pos+1:]
			}
		}
		return fmt.Errorf("rpc 拨打：%s，错误：%s，确保 rpc 服务 %q 已启动",
			server, err.Error(), service)
	}

	c.conn = conn
	return nil
}

func (c *client) buildDialOptions(opts ...ClientOption) []grpc.DialOption {
	var cliOpts ClientOptions
	for _, opt := range opts {
		opt(&cliOpts)
	}

	var options []grpc.DialOption
	if !cliOpts.Secure {
		options = append([]grpc.DialOption(nil), grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if !cliOpts.NonBlock {
		options = append(options, grpc.WithBlock())
	}

	options = append(options,
		WithUnaryClientInterceptors(
			clientinterceptors.UnaryTracingInterceptor,             // 跟踪
			clientinterceptors.DurationInterceptor,                 // 时长
			clientinterceptors.PrometheusInterceptor,               // 统计
			clientinterceptors.BreakerInterceptor,                  // 自动熔断
			clientinterceptors.TimeoutInterceptor(cliOpts.Timeout), // 超时控制
		),
		WithStreamClientInterceptors(
			clientinterceptors.StreamTracingInterceptor,
		),
	)

	return append(options, cliOpts.DialOptions...)
}

// WithDialOption 自定义 ClientOption 的拨号选项。
func WithDialOption(opt grpc.DialOption) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, opt)
	}
}

// WithNonBlock 将拨号设置为非阻塞模式。
func WithNonBlock() ClientOption {
	return func(options *ClientOptions) {
		options.NonBlock = true
	}
}

// WithTimeout 设置 ClientOptions 的超时时长。
func WithTimeout(timeout time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.Timeout = timeout
	}
}

// WithTransportCredentials 确保 gRPC 调用使用给定的凭据进行加密。
func WithTransportCredentials(credentials credentials.TransportCredentials) ClientOption {
	return func(options *ClientOptions) {
		options.Secure = true
		options.DialOptions = append(options.DialOptions, grpc.WithTransportCredentials(credentials))
	}
}

// WithUnaryClientInterceptor 自定义一元客户端拦截器 ClientOptions 的拨号选项。
func WithUnaryClientInterceptor(interceptor grpc.UnaryClientInterceptor) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, WithUnaryClientInterceptors(interceptor))
	}
}

// WithStreamClientInterceptor 自定义流式客户端拦截器 ClientOptions 的拨号选项。
func WithStreamClientInterceptor(interceptor grpc.StreamClientInterceptor) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, WithStreamClientInterceptors(interceptor))
	}
}

// WithStreamClientInterceptors 使用给定的客户端流式拦截器。
func WithStreamClientInterceptors(interceptors ...grpc.StreamClientInterceptor) grpc.DialOption {
	return grpc.WithChainStreamInterceptor(interceptors...)
}

// WithUnaryClientInterceptors 使用给定的客户端一元拦截器。
func WithUnaryClientInterceptors(interceptors ...grpc.UnaryClientInterceptor) grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(interceptors...)
}
