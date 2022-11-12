package logic

import (
	"context"
	"github.com/gotid/god/examples/gateway/hello"
	"github.com/gotid/god/examples/gateway/internal/svc"

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

func (l *PingLogic) Ping(in *hello.Request) (*hello.Response, error) {
	//TODO 此处添加你的逻辑并删除该行

	return &hello.Response{
		Pong: in.Ping,
	}, nil
}
