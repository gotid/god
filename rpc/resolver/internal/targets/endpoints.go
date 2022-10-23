package targets

import (
	"google.golang.org/grpc/resolver"
	"strings"
)

const slashSeparator = "/"

// GetAuthority 获取目标授权。
func GetAuthority(target resolver.Target) string {
	return target.URL.Host
}

// GetEndpoints 返回给定目标的端点。
func GetEndpoints(target resolver.Target) string {
	return strings.Trim(target.URL.Path, slashSeparator)
}
