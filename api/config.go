package api

import (
	"github.com/gotid/god/lib/service"
	"time"
)

type (
	// Config 是一个 http 服务配置项。
	Config struct {
		service.Config
		Host         string `json:",default=0.0.0.0"`
		Port         int
		CertFile     string          `json:",optional"`
		KeyFile      string          `json:",optional"`
		Verbose      bool            `json:",optional"`
		MaxConns     int             `json:",default=10000"`
		MaxBytes     int64           `json:",default=1048576"`
		Timeout      int64           `json:",default=3000"`
		CpuThreshold int64           `json:",default=900,range=[0:1000]"`
		Signature    SignatureConfig `json:",optional"`
	}

	// SignatureConfig 用于服务端签名校验的配置。
	SignatureConfig struct {
		Strict      bool          `json:",default=false"`
		Expire      time.Duration `json:",default=1h"`
		PrivateKeys []PrivateKeyConfig
	}

	// PrivateKeyConfig 用于服务端解密的私钥配置。
	PrivateKeyConfig struct {
		// 信息指纹（与客户端匹配）
		Fingerprint string
		// 指纹对应的私钥文件路径
		KeyFile string
	}
)
