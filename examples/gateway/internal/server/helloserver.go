// Code generated by god. DO NOT EDIT.
// 源文件: hello.proto

package server

import (
	"context"
	hello2 "github.com/gotid/god/examples/gateway/hello"
	"github.com/gotid/god/examples/gateway/internal/logic"
	"github.com/gotid/god/examples/gateway/internal/svc"
)

type HelloServer struct {
	svcCtx *svc.ServiceContext
	hello2.UnimplementedHelloServer
}

func NewHelloServer(svcCtx *svc.ServiceContext) *HelloServer {
	return &HelloServer{
		svcCtx: svcCtx,
	}
}

func (s *HelloServer) Ping(ctx context.Context, in *hello2.Request) (*hello2.Response, error) {
	l := logic.NewPingLogic(ctx, s.svcCtx)
	return l.Ping(in)
}