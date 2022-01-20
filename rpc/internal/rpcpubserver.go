package internal

import (
	"os"
	"strings"

	"github.com/gotid/god/lib/discovery"
	"github.com/gotid/god/lib/netx"
)

const (
	allEths  = "0.0.0.0"
	envPodIp = "POD_IP"
)

// NewPubServer 返回一个新的服务器。
func NewPubServer(etcd discovery.EtcdConf, listenOn string, opts ...ServerOption) (Server, error) {
	registerEtcd := func() error {
		pubListenOn := figureOutListenOn(listenOn)
		var pubOpts []discovery.PublisherOption
		if etcd.HasAccount() {
			pubOpts = append(pubOpts, discovery.WithEtcdAccount(etcd.User, etcd.Pass))
		}
		pubClient := discovery.NewPublisher(etcd.Hosts, etcd.Key, pubListenOn, pubOpts...)
		return pubClient.KeepAlive()
	}
	server := keepAliveServer{
		registerEtcd: registerEtcd,
		Server:       NewRpcServer(listenOn, opts...),
	}

	return server, nil
}

type keepAliveServer struct {
	registerEtcd func() error
	Server
}

func (s keepAliveServer) Start(register RegisterFn) error {
	if err := s.registerEtcd(); err != nil {
		return err
	}

	return s.Server.Start(register)
}

func figureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")

	// 未传监听地址
	if len(fields) == 0 {
		return listenOn
	}

	// 传host且不是0.0.0.0
	host := fields[0]
	if len(host) > 0 && host != allEths {
		return listenOn
	}

	ip := os.Getenv(envPodIp)
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}

	return strings.Join(append([]string{ip}, fields[1:]...), ":")
}
