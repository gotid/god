package logic

import (
	"context"
	"github.com/gotid/god/lib/logx"

	"github.com/gotid/god/examples/http/jwt/internal/svc"
	"github.com/gotid/god/examples/http/jwt/internal/types"
)

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLogic) GetUser(_ *types.GetUserRequest) (resp *types.GetUserResponse, err error) {
	return &types.GetUserResponse{Name: "god"}, nil
}
