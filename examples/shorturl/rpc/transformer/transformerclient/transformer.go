// Code generated by god. DO NOT EDIT.
// 源文件: transformer.proto

//go:generate mockgen -destination ./transformer_mock.go -package transformerclient -source $GOFILE

package transformerclient

import (
	"context"

	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"

	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
)

type (
	ExpandRequest   = transformer.ExpandRequest
	ExpandResponse  = transformer.ExpandResponse
	ShortenRequest  = transformer.ShortenRequest
	ShortenResponse = transformer.ShortenResponse

	Transformer interface {
		Shorten(ctx context.Context, in *ShortenRequest, opts ...grpc.CallOption) (*ShortenResponse, error)
		Expand(ctx context.Context, in *ExpandRequest, opts ...grpc.CallOption) (*ExpandResponse, error)
	}

	defaultTransformer struct {
		cli rpc.Client
	}
)

func NewTransformer(cli rpc.Client) Transformer {
	return &defaultTransformer{
		cli: cli,
	}
}

func (m *defaultTransformer) Shorten(ctx context.Context, in *ShortenRequest, opts ...grpc.CallOption) (*ShortenResponse, error) {
	client := transformer.NewTransformerClient(m.cli.Conn())
	return client.Shorten(ctx, in, opts...)
}

func (m *defaultTransformer) Expand(ctx context.Context, in *ExpandRequest, opts ...grpc.CallOption) (*ExpandResponse, error) {
	client := transformer.NewTransformerClient(m.cli.Conn())
	return client.Expand(ctx, in, opts...)
}
