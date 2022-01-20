package context

import (
	"encoding/json"
	"fmt"
	"time"

	"git.zc0901.com/go/god/lib/wechat/cache"

	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/wechat/util"
)

const urlComponentAccessToken = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"

// ComponentAccessToken 是一个第三方平台访问令牌。
type ComponentAccessToken struct {
	util.WechatError
	AccessToken string `json:"component_access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// SetComponentAccessToken 设置第三方平台访问令牌。
func (ctx *Context) SetComponentAccessToken(verifyTicket string) (*ComponentAccessToken, error) {
	data, err := util.PostJSON(urlComponentAccessToken, g.Map{
		"component_appid":         ctx.AppID,
		"component_appsecret":     ctx.AppSecret,
		"component_verify_ticket": verifyTicket,
	})
	if err != nil {
		return nil, err
	}

	token := &ComponentAccessToken{}
	if err = json.Unmarshal(data, token); err != nil {
		return nil, err
	}

	if token.ErrCode != 0 {
		return nil, fmt.Errorf("SetComponentAccessToken 错误，"+
			"errcode=%d, errmsg=%s", token.ErrCode, token.ErrMsg)
	}

	key := cache.KeyComponentAccessToken(ctx.AppID)
	timeout := time.Duration(token.ExpiresIn-1500) * time.Second
	err = ctx.Cache.Set(key, token.AccessToken, timeout)
	if err != nil {
		return nil, fmt.Errorf("SetComponentAccessToken 错误：%v", err)
	}

	return token, nil
}

// ComponentAccessToken 从缓存中获取第三方平台访问令牌。
func (ctx *Context) ComponentAccessToken() (string, error) {
	key := fmt.Sprintf("component_access_token_%s", ctx.AppID)
	val := ctx.Cache.Get(key)
	if val == nil {
		return "", fmt.Errorf("暂无第三方平台访问令牌缓存，10分钟后再试")
	}
	return val.(string), nil
}
