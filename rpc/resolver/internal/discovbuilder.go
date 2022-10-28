package internal

import (
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/rpc/resolver/internal/targets"
	"google.golang.org/grpc/resolver"
	"strings"
)

type discovBuilder struct{}

func (d *discovBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	hosts := strings.FieldsFunc(targets.GetAuthority(target), func(r rune) bool {
		return r == EndpointSepChar
	})
	sub, err := discov.NewSubscriber(hosts, targets.GetEndpoints(target))
	if err != nil {
		return nil, err
	}

	update := func() {
		var addrs []resolver.Address
		for _, addr := range subset(sub.Values(), subsetSize) {
			addrs = append(addrs, resolver.Address{Addr: addr})
		}
		if err := cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
			logx.Error(err)
		}
	}
	sub.AddListener(update)
	update()

	return &nopResolver{cc: cc}, nil
}

func (d *discovBuilder) Scheme() string {
	return DiscovSchema
}
