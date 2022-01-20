package openplatform

import (
	"git.zc0901.com/go/god/lib/wechat/openplatform/config"
	"git.zc0901.com/go/god/lib/wechat/openplatform/context"
	"git.zc0901.com/go/god/lib/wechat/openplatform/weapp"
)

// OpenPlatform 是微信开放平台相关API结构体。
type OpenPlatform struct {
	*context.Context
}

// New 返回一个新的微信开放平台API。
func New(config *config.Config) *OpenPlatform {
	if config.Cache == nil {
		panic("Cache 未设置")
	}

	ctx := &context.Context{Config: config}
	return &OpenPlatform{ctx}
}

// WeApp 返回一个小程序待开发接口。
func (open *OpenPlatform) WeApp(appID string) *weapp.WeApp {
	return weapp.New(open.Context, appID)
}
