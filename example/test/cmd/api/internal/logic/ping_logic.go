package logic

import (
	"context"

	"github.com/gotid/god/example/test/cmd/rpc/test"

	"github.com/gotid/god/example/test/cmd/api/internal/svc"
	"github.com/gotid/god/example/test/cmd/api/internal/types"

	"github.com/gotid/god/lib/logx"
)

type PingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) PingLogic {
	return PingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PingLogic) Ping(req types.PingReq) (*types.PingReply, error) {
	pingReply, err := l.svcCtx.TestRPC.Ping(l.ctx, &test.PingReq{
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}

	return &types.PingReply{
		Pong: pingReply.Pong,
	}, nil
}
