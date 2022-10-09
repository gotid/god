package redis

import "errors"

var (
	// ErrEmptyHost 是一个表示没设置 redis 主机的错误。
	ErrEmptyHost = errors.New("redis 主机为空")
	// ErrEmptyType 是一个表示没设置 redis 类型的错误。
	ErrEmptyType = errors.New("redis 类型为空")
	// ErrEmptyKey 是一个表示没设置 redis 键名的错误。
	ErrEmptyKey = errors.New("redis 键名为空")
)

type (
	// Config 是一个 redis 配置。
	Config struct {
		Host string
		Type string `json:",default=node,options=[node,cluster]"`
		Pass string `json:",optional"`
		Tls  bool   `json:",optional"`
	}

	// KeyConfig 是一个基于给定键的 redis 配置。
	KeyConfig struct {
		Config
		Key string `json:",optional"`
	}
)

// NewRedis 基于配置返回一个 Redis 节点实例。
func (c Config) NewRedis() *Redis {
	var opts []Option
	if c.Type == ClusterType {
		opts = append(opts, WithCluster())
	}
	if len(c.Pass) > 0 {
		opts = append(opts, WithPass(c.Pass))
	}
	if c.Tls {
		opts = append(opts, WithTLS())
	}

	return New(c.Host, opts...)
}

// Validate 验证 Config 是否正确。
func (c Config) Validate() error {
	if len(c.Host) == 0 {
		return ErrEmptyHost
	}

	if len(c.Type) == 0 {
		return ErrEmptyType
	}

	return nil
}

// Validate 验证 KeyConfig 是否正确。
func (kc KeyConfig) Validate() error {
	if err := kc.Config.Validate(); err != nil {
		return err
	}

	if len(kc.Key) == 0 {
		return ErrEmptyKey
	}

	return nil
}
