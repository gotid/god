package weapp

import (
	"git.zc0901.com/go/god/lib/wechat/openplatform/context"
)

// WeApp 是一个小程序待开发接口。
type WeApp struct {
	AppID   string
	context *context.Context
}

// New 返回一个小程序待开发接口。
func New(ctx *context.Context, appID string) *WeApp {
	return &WeApp{
		AppID:   appID,
		context: ctx,
	}
}
