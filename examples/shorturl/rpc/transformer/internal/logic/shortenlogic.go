package logic

import (
	"context"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/internal/svc"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/model"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"
	"github.com/gotid/god/lib/hash"

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
	key := hash.Md5Hex([]byte(in.Url))[:6]
	_, err := l.svcCtx.Model.Insert(l.ctx, &model.Shorturl{
		Shorten: key,
		Url:     in.Url,
	})
	if err != nil {
		return nil, err
	}

	return &transformer.ShortenResponse{
		Shorten: key,
	}, nil
}
