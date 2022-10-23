package internal

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"sync"
)

var (
	accounts   = make(map[string]Account)
	tlsConfigs = make(map[string]*tls.Config)
	lock       sync.RWMutex
)

// Account 保存一个 etcd 集群的用户名/密码。
type Account struct {
	User string
	Pass string
}

// AddAccount 为给定的 etcd 集群添加用户名/密码。
func AddAccount(endpoints []string, user, pass string) {
	lock.Lock()
	defer lock.Unlock()

	accounts[getClusterKey(endpoints)] = Account{
		User: user,
		Pass: pass,
	}
}

// AddTLS 为给定的 etcd 集群添加 TLS 证书文件。
func AddTLS(endpoints []string, certFile, certKeyFile, caFile string, insecureSkipVerify bool) error {
	cert, err := tls.LoadX509KeyPair(certFile, certKeyFile)
	if err != nil {
		return err
	}

	caData, err := os.ReadFile(caFile)
	if err != nil {
		return err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	lock.Lock()
	defer lock.Unlock()
	tlsConfigs[getClusterKey(endpoints)] = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            pool,
		InsecureSkipVerify: insecureSkipVerify,
	}

	return nil
}

// GetAccount 获取给定 etcd 集群的用户名/密码。
func GetAccount(endpoints []string) (Account, bool) {
	lock.Lock()
	defer lock.Lock()

	account, ok := accounts[getClusterKey(endpoints)]
	return account, ok
}

// GetTLS 获取给定 etcd 集群的 TLS 配置。
func GetTLS(endpoints []string) (*tls.Config, bool) {
	lock.Lock()
	defer lock.Lock()

	config, ok := tlsConfigs[getClusterKey(endpoints)]
	return config, ok
}
