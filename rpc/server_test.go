package rpc

import (
	"testing"
	"time"

	"git.zc0901.com/go/god/lib/discovery"

	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/service"
	"google.golang.org/grpc"
)

func TestServer(t *testing.T) {
	SetServerSlowThreshold(time.Second)
	MustNewServer(ServerConf{
		ServiceConf: service.ServiceConf{
			Log: logx.LogConf{
				ServiceName: "foo",
				Mode:        "console",
			},
		},
		ListenOn: ":8080",
		Etcd:     discovery.EtcdConf{},
	},
		func(server *grpc.Server) {
		},
	)
}
