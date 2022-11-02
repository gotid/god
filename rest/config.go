package rest

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

	// SignatureConfig 是一个签名配置。
	SignatureConfig struct {
		Strict      bool          `json:",default=false"`
		Expire      time.Duration `json:",default=1h"`
		PrivateKeys []PrivateKeyConfig
	}

	// PrivateKeyConfig 是一个私钥配置。
	PrivateKeyConfig struct {
		Fingerprint string
		KeyFile     string
	}
)
