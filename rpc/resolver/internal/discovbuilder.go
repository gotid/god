package internal

import (
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/rpc/resolver/internal/targets"
	"google.golang.org/grpc/resolver"
	"strings"
)

type discovBuilder struct{}

func (d *discovBuilder) Build(target resolver.Target, conn resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	hosts := strings.FieldsFunc(targets.GetAuthority(target), func(r rune) bool {
		return r == EndpointSepChar
	})
	sub, err := discov.NewSubscriber(hosts, targets.GetEndpoints(target))
	if err != nil {
		return nil, err
	}

	update := func() {
		var addrs []resolver.Address
		for _, val := range subset(sub.Values(), subsetSize) {
			addrs = append(addrs, resolver.Address{Addr: val})
		}
		if err := conn.UpdateState(resolver.State{Addresses: addrs}); err != nil {
			logx.Error(err)
		}
	}
	sub.AddListener(update)
	update()

	return &nopResolver{conn: conn}, nil
}

func (d *discovBuilder) Scheme() string {
	return DiscovSchema
}
