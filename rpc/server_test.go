package rpc

import (
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/rpc/internal"
	"github.com/gotid/god/rpc/internal/serverinterceptors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestServer_AddInterceptors(t *testing.T) {
	server := new(mockedServer)
	err := setupInterceptors(server, ServerConfig{
		Auth: true,
		Redis: redis.KeyConfig{
			Config: redis.Config{
				Host: "any",
				Type: redis.NodeType,
			},
			Key: "",
		},
		Timeout:      100,
		CpuThreshold: 10,
	}, new(stat.Metrics))
	assert.Nil(t, err)
	assert.Equal(t, 3, len(server.unaryInterceptors))
	assert.Equal(t, 1, len(server.streamInterceptors))
}

func TestServer(t *testing.T) {
	DontLogContentForMethod("foo")
	SetServerSlowThreshold(time.Second)
	svr := MustNewServer(ServerConfig{
		Config:        service.Config{},
		ListenOn:      "localhost:8080",
		Etcd:          discov.EtcdConfig{},
		Auth:          false,
		Redis:         redis.KeyConfig{},
		StrictControl: false,
		Timeout:       0,
		CpuThreshold:  0,
	}, func(server *grpc.Server) {})
	svr.AddOptions(grpc.ConnectionTimeout(time.Hour))
	svr.AddUnaryInterceptors(serverinterceptors.UnaryCrashInterceptor)
	svr.AddStreamInterceptors(serverinterceptors.StreamCrashInterceptor)
	go svr.Start()
	svr.Stop()
}

func TestServerError(t *testing.T) {
	_, err := NewServer(ServerConfig{
		Config: service.Config{
			Log: logx.Config{
				ServiceName: "foo",
				Mode:        "console",
			},
		},
		ListenOn: "localhost:8080",
		Etcd: discov.EtcdConfig{
			Hosts: []string{"localhost"},
		},
		Auth:  true,
		Redis: redis.KeyConfig{},
	}, func(server *grpc.Server) {})
	assert.NotNil(t, err)
}

func TestServer_HasEtcd(t *testing.T) {
	svr := MustNewServer(ServerConfig{
		Config: service.Config{
			Log: logx.Config{
				ServiceName: "foo",
				Mode:        "console",
			},
		},
		ListenOn: "localhost:8080",
		Etcd: discov.EtcdConfig{
			Hosts: []string{"notexist"},
			Key:   "any",
		},
		Redis: redis.KeyConfig{},
	}, func(server *grpc.Server) {
	})
	svr.AddOptions(grpc.ConnectionTimeout(time.Hour))
	svr.AddUnaryInterceptors(serverinterceptors.UnaryCrashInterceptor)
	svr.AddStreamInterceptors(serverinterceptors.StreamCrashInterceptor)
	go svr.Start()
	svr.Stop()
}

func TestServer_StartFailed(t *testing.T) {
	svr := MustNewServer(ServerConfig{
		Config: service.Config{
			Log: logx.Config{
				ServiceName: "foo",
				Mode:        "console",
			},
		},
		ListenOn: "localhost:aaa",
	}, func(server *grpc.Server) {
	})

	assert.Panics(t, svr.Start)
}

type mockedServer struct {
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

func (m *mockedServer) AddOptions(_ ...grpc.ServerOption) {

}

func (m *mockedServer) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	m.streamInterceptors = append(m.streamInterceptors, interceptors...)
}

func (m *mockedServer) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	m.unaryInterceptors = append(m.unaryInterceptors, interceptors...)
}

func (m *mockedServer) SetName(_ string) {

}

func (m *mockedServer) Start(_ internal.RegisterFn) error {
	return nil
}
