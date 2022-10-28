package logic

import (
	"context"

	"github.com/gotid/god/tools/god/internal/svc"
	"github.com/gotid/god/tools/god/pb/test"

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

func (l *PingLogic) Ping(in *test.Request) (*test.Response, error) {
	// todo: 此处添加你的逻辑并删除该行

	return &test.Response{}, nil
}
