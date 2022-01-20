// Code generated by god. DO NOT EDIT!
// Source: graceful.proto

//go:generate mockgen -destination ./graceful_service_mock.go -package gracefulservice -source $GOFILE

package gracefulservice

import (
	"context"

	"github.com/gotid/god/example/graceful/dns/rpc/graceful"

	"github.com/gotid/god/rpc"
)

type (
	Request  = graceful.Request
	Response = graceful.Response

	GracefulService interface {
		Grace(ctx context.Context, in *Request) (*Response, error)
	}

	defaultGracefulService struct {
		cli rpc.Client
	}
)

func NewGracefulService(cli rpc.Client) GracefulService {
	return &defaultGracefulService{
		cli: cli,
	}
}

func (m *defaultGracefulService) Grace(ctx context.Context, in *Request) (*Response, error) {
	client := graceful.NewGracefulServiceClient(m.cli.Conn())
	return client.Grace(ctx, in)
}
