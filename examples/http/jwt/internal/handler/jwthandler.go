package handler

import (
	"github.com/gotid/god/examples/http/jwt/internal/logic"
	"github.com/gotid/god/examples/http/jwt/internal/svc"
	"github.com/gotid/god/examples/http/jwt/internal/types"
	"net/http"

	"github.com/gotid/god/api/httpx"
)

func JwtHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.JwtTokenRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewJwtLogic(r.Context(), svcCtx)
		resp, err := l.Jwt(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
