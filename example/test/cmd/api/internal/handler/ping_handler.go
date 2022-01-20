package handler

import (
	"net/http"

	"github.com/gotid/god/example/test/cmd/api/internal/logic"
	"github.com/gotid/god/example/test/cmd/api/internal/svc"
	"github.com/gotid/god/example/test/cmd/api/internal/types"

	"github.com/gotid/god/api/httpx"
)

func PingHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewPingLogic(r.Context(), ctx)
		resp, err := l.Ping(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
