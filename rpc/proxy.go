package rpc

import (
	"context"
	"sync"

	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/rpc/internal"
	"github.com/gotid/god/rpc/internal/auth"
	"google.golang.org/grpc"
)

type RpcProxy struct {
	backend     string // 代理对应的后端服务器地址
	clients     map[string]Client
	options     []internal.ClientOption
	sharedCalls syncx.SingleFlight
	lock        sync.Mutex
}

func (p *RpcProxy) TakeConn(ctx context.Context) (*grpc.ClientConn, error) {
	cred := auth.ParseCredential(ctx)
	key := cred.App + "/" + cred.Token
	val, _, err := p.sharedCalls.Do(key, func() (interface{}, error) {
		p.lock.Lock()
		client, ok := p.clients[key]
		p.lock.Unlock()
		if ok {
			return client, nil
		}

		opts := append(p.options, WithDialOption(grpc.WithPerRPCCredentials(&auth.Credential{
			App:   cred.App,
			Token: cred.Token,
		})))
		client, err := NewClientWithTarget(p.backend, opts...)
		if err != nil {
			return nil, err
		}

		p.lock.Lock()
		p.clients[key] = client
		p.lock.Unlock()
		return client, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(Client).Conn(), nil
}

func NewProxy(backend string, opts ...internal.ClientOption) *RpcProxy {
	return &RpcProxy{
		backend:     backend,
		clients:     make(map[string]Client),
		options:     opts,
		sharedCalls: syncx.NewSingleFlight(),
	}
}
