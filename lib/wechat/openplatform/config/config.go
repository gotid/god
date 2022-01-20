package config

import "git.zc0901.com/go/god/lib/wechat/cache"

// Config 微信开发平台配置项。
type Config struct {
	AppID          string `json:"app_id"`           // 开放平台 APPID
	AppSecret      string `json:"app_secret"`       // 开放平台 AppSecret
	Token          string `json:"token"`            // 开放平台——授权后实现业务——消息校验Token
	EncodingAESKey string `json:"encoding_aes_key"` // 开放平台——授权后实现业务——消息加解密Key
	Cache          cache.Cache
}
