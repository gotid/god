package logic

import (
	"context"
	"time"

	"github.com/gotid/god/example/test/cmd/rpc/internal/svc"
	"github.com/gotid/god/example/test/cmd/rpc/test"

	"github.com/gotid/god/lib/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PingLogic) Ping(req *test.PingReq) (*test.PingReply, error) {
	time.Sleep(5 * time.Second)
	return &test.PingReply{
		Pong: req.Name,
	}, nil
}
