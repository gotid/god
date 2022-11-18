package logic

import (
	"context"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"

	"github.com/gotid/god/examples/shorturl/api/internal/svc"
	"github.com/gotid/god/examples/shorturl/api/internal/types"

	"github.com/gotid/god/lib/logx"
)

type ExpandLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExpandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExpandLogic {
	return &ExpandLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExpandLogic) Expand(req *types.ExpandRequest) (resp *types.ExpandResponse, err error) {
	reply, err := l.svcCtx.Transformer.Expand(l.ctx, &transformer.ExpandRequest{
		Shorten: req.Shorten,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.ExpandResponse{
		Url: reply.Url,
	}

	return
}
