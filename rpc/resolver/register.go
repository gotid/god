package resolver

import "github.com/gotid/god/rpc/resolver/internal"

// Register 注册 rpc 定义的方案。
// 保存在单独包中，以便第三方手动注册。
func Register() {
	internal.RegisterResolver()
}
