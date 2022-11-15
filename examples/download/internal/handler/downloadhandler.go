package handler

import (
	"net/http"
	"os"

	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/examples/download/internal/svc"
	"github.com/gotid/god/examples/download/internal/types"
)

func DownloadHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Request
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		body, err := os.ReadFile(req.File)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		w.Write(body)
	}
}
