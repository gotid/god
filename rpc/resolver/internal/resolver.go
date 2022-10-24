package internal

import (
	"fmt"
	"google.golang.org/grpc/resolver"
)

const (
	// DirectScheme 代表直连方案。
	DirectScheme = "direct"
	// DiscovSchema 代表 discov 方案。
	DiscovSchema = "discov"
	// EtcdSchema 代表 etcd 方案。
	EtcdSchema = "etcd"
	// EndpointSepChar 是端点中的分隔符字符。
	EndpointSepChar = ','

	subsetSize = 32
)

var (
	// EndpointSep 是端点中的分隔符字符串。
	EndpointSep = fmt.Sprintf("%c", EndpointSepChar)

	directResolverBuilder directBuilder
	discovResolverBuilder discovBuilder
	etcdResolverBuilder   etcdBuilder
)

// RegisterResolver 注册服务直连和服务发现方案到解析器。
func RegisterResolver() {
	resolver.Register(&directResolverBuilder)
	resolver.Register(&discovResolverBuilder)
	resolver.Register(&etcdResolverBuilder)
}

type nopResolver struct {
	conn resolver.ClientConn
}

func (r *nopResolver) Close() {}

func (r *nopResolver) ResolveNow(_ resolver.ResolveNowOptions) {}
