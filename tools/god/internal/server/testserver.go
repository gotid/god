// 代码由 god 生成，不要修改。
// 源文件: test.proto

package server

import (
	"context"

	"github.com/gotid/god/tools/god/internal/logic"
	"github.com/gotid/god/tools/god/internal/svc"
	"github.com/gotid/god/tools/god/pb/test"
)

type TestServer struct {
	svcCtx *svc.ServiceContext
	test.UnimplementedTestServer
}

func NewTestServer(svcCtx *svc.ServiceContext) *TestServer {
	return &TestServer{
		svcCtx: svcCtx,
	}
}

func (s *TestServer) Ping(ctx context.Context, in *test.Request) (*test.Response, error) {
	l := logic.NewPingLogic(ctx, s.svcCtx)
	return l.Ping(in)
}
