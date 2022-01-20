// Code generated by god. DO NOT EDIT!
// Source: test.proto

package server

import (
	"context"

	"github.com/gotid/god/example/test/cmd/rpc/internal/logic"
	"github.com/gotid/god/example/test/cmd/rpc/internal/svc"
	"github.com/gotid/god/example/test/cmd/rpc/test"
)

type TestServer struct {
	svcCtx *svc.ServiceContext
}

func NewTestServer(svcCtx *svc.ServiceContext) *TestServer {
	return &TestServer{
		svcCtx: svcCtx,
	}
}

func (s *TestServer) Ping(ctx context.Context, req *test.PingReq) (*test.PingReply, error) {
	l := logic.NewPingLogic(ctx, s.svcCtx)
	return l.Ping(req)
}
