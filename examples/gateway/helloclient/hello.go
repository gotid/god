// Code generated by god. DO NOT EDIT.
// 源文件: hello.proto

package helloclient

import (
	"context"
	hello2 "github.com/gotid/god/examples/gateway/hello"

	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
)

type (
	Request  = hello2.Request
	Response = hello2.Response

	Hello interface {
		Ping(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	}

	defaultHello struct {
		cli rpc.Client
	}
)

func NewHello(cli rpc.Client) Hello {
	return &defaultHello{
		cli: cli,
	}
}

func (m *defaultHello) Ping(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	client := hello2.NewHelloClient(m.cli.Conn())
	return client.Ping(ctx, in, opts...)
}
