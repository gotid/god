package resolver

import (
	"strings"

	"github.com/gotid/god/lib/discovery"
	"google.golang.org/grpc/resolver"
)

// discoveryBuilder 服务直连构建器
type discoveryBuilder struct{}

// Build 构建服务发现解析器
func (b *discoveryBuilder) Build(target resolver.Target, conn resolver.ClientConn, options resolver.BuildOptions) (
	resolver.Resolver, error) {
	endpoints := strings.FieldsFunc(target.Authority, func(r rune) bool {
		return r == EndpointSepChar
	})
	subscriber, err := discovery.NewSubscriber(endpoints, target.Endpoint)
	if err != nil {
		return nil, err
	}

	update := func() {
		var addrs []resolver.Address
		for _, addr := range subset(subscriber.Values(), subsetSize) {
			addrs = append(addrs, resolver.Address{Addr: addr})
		}
		conn.UpdateState(resolver.State{Addresses: addrs})
	}
	subscriber.AddListener(update)
	update()

	return &nopResolver{conn: conn}, nil
}

// Schema 服务发现模式
func (b *discoveryBuilder) Scheme() string {
	return DiscoverySchema
}
