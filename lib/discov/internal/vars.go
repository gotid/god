package internal

import "time"

const (
	// Delimiter 是一个分割 etcd 路径的分隔符。
	Delimiter = '/'

	endpointsSeparator = ","
	requestTimeout     = 3 * time.Second
	dialTimeout        = 5 * time.Second
	dailyKeepAliveTime = 5 * time.Second
	coolDownInterval   = 1 * time.Second
	autoSyncInterval   = 1 * time.Minute
)

var (
	// RequestTimeout 是请求超时时长，默认为3秒。
	RequestTimeout = requestTimeout
	// DialTimeout 是拨号超时时长。
	DialTimeout = dialTimeout
)
