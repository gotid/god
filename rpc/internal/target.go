package internal

import (
	"fmt"
	"strings"

	"github.com/gotid/god/rpc/internal/resolver"
)

func BuildDirectTarget(endpoints []string) string {
	return fmt.Sprintf("%s:///%s", resolver.DirectSchema,
		strings.Join(endpoints, resolver.EndpointsSep))
}

func BuildDiscoveryTarget(endpoints []string, key string) string {
	return fmt.Sprintf("%s://%s/%s", resolver.DiscoverySchema,
		strings.Join(endpoints, resolver.EndpointsSep), key)
}
