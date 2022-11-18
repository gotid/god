package logic

import (
	"context"

	"github.com/gotid/god/examples/shorturl/rpc/transformer/internal/svc"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"

	"github.com/gotid/god/lib/logx"
)

type ShortenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewShortenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShortenLogic {
	return &ShortenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ShortenLogic) Shorten(in *transformer.ShortenRequest) (*transformer.ShortenResponse, error) {
	//TODO 此处添加你的逻辑并删除该行

	return &transformer.ShortenResponse{}, nil
}
