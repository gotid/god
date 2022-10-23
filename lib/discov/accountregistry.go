package discov

import "github.com/gotid/god/lib/discov/internal"

// RegisterAccount 将用户名/密码注册到给定的 etcd 集群。
func RegisterAccount(endpoints []string, user, pass string) {
	internal.AddAccount(endpoints, user, pass)
}

// RegisterTLS 将证书注册到给定的 etcd 集群。
func RegisterTLS(endpoints []string, certFile, certKeyFile, caFile string,
	insecureSkipVerify bool) error {
	return internal.AddTLS(endpoints, certFile, certKeyFile, caFile, insecureSkipVerify)
}
