package gateway

import (
	"github.com/gotid/god/api"
	"github.com/gotid/god/rpc"
	"time"
)

type (
	// Config 是关网配置。
	Config struct {
		api.Config
		Upstreams []Upstream
		Timeout   time.Duration `json:",default=5s"`
	}

	// Upstream 是上游 grpc 配置。
	Upstream struct {
		// Name 是上游的名称
		Name string `json:",optional"`
		// Grpc 是上游的目标
		Grpc rpc.ClientConfig
		// ProtoSets 是 proto 文件列表，形如 [hello.pb]。
		// 如果你的 proto 文件导入了其他 proto，逆序编写多文件切片，如：
		// [hello.pb, common.pb]。
		ProtoSets []string `json:",optional"`
		// Mappings 是网关路由和上游 rpc 方法之间的映射。
		// 如果再 rpc 方法中添加了注释，请将其保留为空。
		Mappings []RouteMapping `json:",optional"`
	}

	// RouteMapping 是网关路由和上游 rpc 方法之间的映射。
	RouteMapping struct {
		// Method 是 HTTP 方法，如 GET，POST，PUT，DELETE
		Method string
		// Path 是 HTTP 路径。
		Path string
		// RpcPath 是 rpc 方法，格式如 package.service/method
		RpcPath string
	}
)
