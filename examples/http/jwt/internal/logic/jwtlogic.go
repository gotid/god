package logic

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gotid/god/examples/http/jwt/internal/svc"
	"github.com/gotid/god/examples/http/jwt/internal/types"
	"github.com/gotid/god/lib/logx"
	"time"
)

type JwtLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJwtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JwtLogic {
	return &JwtLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JwtLogic) Jwt(_ *types.JwtTokenRequest) (resp *types.JwtTokenResponse, err error) {
	accessExpire := l.svcCtx.Config.JwtAuth.AccessExpire

	now := time.Now().Unix()
	accessToken, err := l.GenToken(now, l.svcCtx.Config.JwtAuth.AccessSecret, nil, accessExpire)
	if err != nil {
		return nil, err
	}

	return &types.JwtTokenResponse{
		AccessToken:  accessToken,
		AccessExpire: now + accessExpire,
		RefreshAfter: now + accessExpire/2,
	}, nil
}

func (l *JwtLogic) GenToken(iat int64, secretKey string, payloads map[string]interface{}, seconds int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	for k, v := range payloads {
		claims[k] = v
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(secretKey))
}
