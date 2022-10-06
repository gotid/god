package discov

import "errors"

// EtcdConfig 是基于 etcd 给定键的配置项。
type EtcdConfig struct {
	Hosts              []string
	Key                string
	User               string `json:",optional"`
	Pass               string `json:",optional"`
	CertFile           string `json:",optional"`
	CertKeyFile        string `json:",optional=CertFile"`
	CACertFile         string `json:",optional=CertFile"`
	InsecureSkipVerify bool   `json:",optional"`
}

var (
	// 代表 etcd 主机为空。
	errEmptyEtcdHosts = errors.New("etcd 主机不可为空")
	// 代表 etcd 键为空。
	errEmptyEtcdKey = errors.New("etcd 键不可为空")
)

// Validate 验证配置是否正确。
func (c EtcdConfig) Validate() error {
	if len(c.Hosts) == 0 {
		return errEmptyEtcdHosts
	} else if len(c.Key) == 0 {
		return errEmptyEtcdKey
	} else {
		return nil
	}
}

// HasAccount 判断配置中是否提供了用户和密码。
func (c EtcdConfig) HasAccount() bool {
	return len(c.User) > 0 && len(c.Pass) > 0
}

// HasTLS 判断配置中是否提供了完备的 TLS 证书文件。
func (c EtcdConfig) HasTLS() bool {
	return len(c.CertFile) > 0 && len(c.CertKeyFile) > 0 && len(c.CACertFile) > 0
}
