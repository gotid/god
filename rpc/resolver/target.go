package resolver

import (
	"fmt"
	"github.com/gotid/god/rpc/resolver/internal"
	"strings"
)

// BuildDirectTarget 返回给定端点和直连方案的字符串表示形式。
func BuildDirectTarget(endpoints []string) string {
	return fmt.Sprintf("%s:///%s", internal.DirectScheme,
		strings.Join(endpoints, internal.EndpointSep))
}

// BuildDiscovTarget 返回给定端点和 discov 方案的字符串表示形式。
func BuildDiscovTarget(endpoints []string, key string) string {
	return fmt.Sprintf("%s://%s/%s", internal.EtcdSchema,
		strings.Join(endpoints, internal.EndpointSep), key)
}
