package handler

import (
	"net/http"

	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/examples/upload/internal/logic"
	"github.com/gotid/god/examples/upload/internal/svc"
)

func UploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUploadLogic(r, svcCtx)
		resp, err := l.Upload()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
