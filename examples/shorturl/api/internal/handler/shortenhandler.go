package handler

import (
	"net/http"

	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/examples/shorturl/api/internal/logic"
	"github.com/gotid/god/examples/shorturl/api/internal/svc"
	"github.com/gotid/god/examples/shorturl/api/internal/types"
)

func ShortenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ShortenRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewShortenLogic(r.Context(), svcCtx)
		resp, err := l.Shorten(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
