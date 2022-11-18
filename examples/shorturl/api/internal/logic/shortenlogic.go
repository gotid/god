package logic

import (
	"context"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"

	"github.com/gotid/god/examples/shorturl/api/internal/svc"
	"github.com/gotid/god/examples/shorturl/api/internal/types"

	"github.com/gotid/god/lib/logx"
)

type ShortenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShortenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShortenLogic {
	return &ShortenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShortenLogic) Shorten(req *types.ShortenRequest) (resp *types.ShortenResponse, err error) {
	reply, err := l.svcCtx.Transformer.Shorten(l.ctx, &transformer.ShortenRequest{
		Url: req.Url,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.ShortenResponse{
		Shorten: reply.Shorten,
	}

	return
}
