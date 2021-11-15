package rpc

import (
	"log"
	"time"

	"git.zc0901.com/go/god/lib/load"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/stat"
	"git.zc0901.com/go/god/rpc/internal"
	"git.zc0901.com/go/god/rpc/internal/auth"
	"git.zc0901.com/go/god/rpc/internal/serverinterceptors"
	"google.golang.org/grpc"
)

type RpcServer struct {
	server   internal.Server
	register internal.RegisterFn
}

// MustNewServer 返回一个新的 RpcServer，有错退出。
func MustNewServer(sc ServerConf, register internal.RegisterFn) *RpcServer {
	server, err := NewServer(sc, register)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

// NewServer 返回一个新的 RpcServer。
func NewServer(c ServerConf, register internal.RegisterFn) (*RpcServer, error) {
	var err error

	// 验证服务端配置
	if err = c.Validate(); err != nil {
		return nil, err
	}

	// 新建监控指标，以监听端口作为监听名称
	metrics := stat.NewMetrics(c.ListenOn)

	// 新建内部RPC服务器
	var server internal.Server
	serverOptions := []internal.ServerOption{
		internal.WithMetrics(metrics),
		internal.WithMaxRetries(c.MaxRetries),
	}

	if c.HasEtcd() {
		server, err = internal.NewPubServer(c.Etcd, c.ListenOn, serverOptions...)
		if err != nil {
			return nil, err
		}
	} else {
		server = internal.NewRpcServer(c.ListenOn, serverOptions...)
	}

	server.SetName(c.Name)
	if err = setupInterceptors(server, c, metrics); err != nil {
		return nil, err
	}

	// 新建对外RPC服务器
	rpcServer := &RpcServer{
		server:   server,
		register: register,
	}
	if err = c.Setup(); err != nil {
		return nil, err
	}

	return rpcServer, nil
}

// AddOptions 添加 grpc.ServerOption 选项。
func (rs *RpcServer) AddOptions(options ...grpc.ServerOption) {
	rs.server.AddOptions(options...)
}

// AddUnaryInterceptors 添加指定的一元拦截器。
func (rs *RpcServer) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	rs.server.AddUnaryInterceptors(interceptors...)
}

// AddStreamInterceptors 添加指定的流式拦截器。
func (rs *RpcServer) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	rs.server.AddStreamInterceptors(interceptors...)
}

// Start 启动 RpcServer。
// 默认平滑关闭。
// 使用 proc.SetTimeToForceQuit 可自定义平滑关闭周期。
func (rs *RpcServer) Start() {
	if err := rs.server.Start(rs.register); err != nil {
		logx.Error(err)
		panic(err)
	}
}

// Stop 停止 RpcServer。
func (rs *RpcServer) Stop() {
	logx.Close()
}

// SetServerSlowThreshold 设置服务端慢调用阈值。
func SetServerSlowThreshold(duration time.Duration) {
	serverinterceptors.SetSlowThreshold(duration)
}

func setupInterceptors(server internal.Server, c ServerConf, metrics *stat.Metrics) error {
	// 自动降载（负载泄流拦截器）
	if c.CpuThreshold > 0 {
		shedder := load.NewAdaptiveShedder(load.WithCpuThreshold(c.CpuThreshold))
		server.AddUnaryInterceptors(serverinterceptors.UnaryShedderInterceptor(shedder, metrics))
	}

	// 超时控制（超时拦截器）
	if c.Timeout > 0 {
		server.AddUnaryInterceptors(serverinterceptors.UnaryTimeoutInterceptor(
			time.Duration(c.Timeout) * time.Millisecond))
	}

	// 调用鉴权（鉴权拦截器）
	if c.Auth {
		authenticator, err := auth.NewAuthenticator(c.Redis.NewRedis(), c.Redis.Key, c.StrictControl)
		if err != nil {
			return err
		}

		server.AddStreamInterceptors(serverinterceptors.StreamAuthorizeInterceptor(authenticator))
		server.AddUnaryInterceptors(serverinterceptors.UnaryAuthorizeInterceptor(authenticator))
	}

	return nil
}
