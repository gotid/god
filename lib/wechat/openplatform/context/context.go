package context

import "git.zc0901.com/go/god/lib/wechat/openplatform/config"

// Context 微信开放平台上下文。
type Context struct {
	*config.Config
	// wechat.AccessToken
}
