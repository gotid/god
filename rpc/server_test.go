package rpc

import (
	"testing"
	"time"

	"github.com/gotid/god/lib/discovery"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/service"
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
