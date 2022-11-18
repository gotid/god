package logic

import (
	"context"
	"github.com/gotid/god/examples/shorturl/errorx"

	"github.com/gotid/god/examples/shorturl/rpc/transformer/internal/svc"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"

	"github.com/gotid/god/lib/logx"
)

type ExpandLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExpandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExpandLogic {
	return &ExpandLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ExpandLogic) Expand(in *transformer.ExpandRequest) (*transformer.ExpandResponse, error) {
	resp, err := l.svcCtx.Model.FindOne(l.ctx, in.Shorten)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	return &transformer.ExpandResponse{
		Url: resp.Url,
	}, nil
}
