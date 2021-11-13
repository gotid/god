package discovery

import "errors"

type EtcdConf struct {
	Hosts []string // etcd监听的ip数组
	Key   string   // rpc注册key
	User  string   `json:",optional"`
	Pass  string   `json:",optional"`
}

// HasAccount 返回是否提供了Etcd账号。
func (c EtcdConf) HasAccount() bool {
	return len(c.User) > 0 && len(c.Pass) > 0
}

// Validate 验证Etcd配置项
func (c EtcdConf) Validate() error {
	if len(c.Hosts) == 0 {
		return errors.New("未配置用于服务发现的 Etcd Hosts")
	} else if len(c.Key) == 0 {
		return errors.New("未配置用于服务发现的 Etcd Key")
	} else {
		return nil
	}
}
