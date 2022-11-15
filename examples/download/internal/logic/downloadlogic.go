package logic

import (
	"context"

	"github.com/gotid/god/examples/download/internal/svc"
	"github.com/gotid/god/examples/download/internal/types"

	"github.com/gotid/god/lib/logx"
)

type DownloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadLogic {
	return &DownloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadLogic) Download(req *types.Request) (resp *types.Response, err error) {
	//TODO 添加你的逻辑并删除此行

	return
}
