package internal

import (
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/netx"
	"os"
	"strings"
)

const (
	allEths  = "0.0.0.0"
	envPodId = "POD_IP"
)

type keepAliveServer struct {
	registerEtcd func() error
	Server
}

func (s keepAliveServer) Start(fn RegisterFn) error {
	if err := s.registerEtcd(); err != nil {
		return err
	}

	return s.Server.Start(fn)
}

// NewPubServer 返回一个基于 etcd 的 rpc 服务。
func NewPubServer(etcd discov.EtcdConfig, listenOn string, opts ...ServerOption) (Server, error) {
	registerEtcd := func() error {
		pubListenOn := figureOutListenOn(listenOn)
		var pubOpts []discov.PubOption
		if etcd.HasAccount() {
			pubOpts = append(pubOpts, discov.WithPubEtcdAccount(etcd.User, etcd.Pass))
		}
		if etcd.HasTLS() {
			pubOpts = append(pubOpts, discov.WithPubEtcdTLS(etcd.CertFile, etcd.CertKeyFile,
				etcd.CACertFile, etcd.InsecureSkipVerify))
		}
		pubClient := discov.NewPublisher(etcd.Hosts, etcd.Key, pubListenOn, pubOpts...)
		return pubClient.KeepAlive()
	}
	svr := keepAliveServer{
		registerEtcd: registerEtcd,
		Server:       NewServer(listenOn, opts...),
	}

	return svr, nil
}

func figureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")
	if len(fields) == 0 {
		return listenOn
	}

	host := fields[0]
	if len(host) > 0 && host != allEths {
		return listenOn
	}

	ip := os.Getenv(envPodId)
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}

	return strings.Join(append([]string{ip}, fields[1:]...), ":")
}
