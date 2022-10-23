package rpc

import (
	"context"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/rpc/internal/auth"
	"google.golang.org/grpc"
	"sync"
)

// Proxy 是一个 rpc 代理。
type Proxy struct {
	backend      string
	clients      map[string]Client
	options      []ClientOption
	singleFlight syncx.SingleFlight
	lock         sync.Mutex
}

// NewProxy 返回一个 rpc 代理 Proxy。
func NewProxy(backend string, opts ...ClientOption) *Proxy {
	return &Proxy{
		backend:      backend,
		clients:      make(map[string]Client),
		options:      opts,
		singleFlight: syncx.NewSingleFlight(),
	}
}

// TakeConn 返回一个 grpc.ClientConn。
func (p *Proxy) TakeConn(ctx context.Context) (*grpc.ClientConn, error) {
	cred := auth.ParseCredential(ctx)
	key := cred.App + "/" + cred.Token
	val, err := p.singleFlight.Do(key, func() (interface{}, error) {
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
