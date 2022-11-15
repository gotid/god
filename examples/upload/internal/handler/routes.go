// Code generated by god. DO NOT EDIT.
package handler

import (
	"net/http"

	"github.com/gotid/god/examples/upload/internal/svc"

	"github.com/gotid/god/api"
)

func RegisterHandlers(server *api.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]api.Route{
			{
				Method:  http.MethodPost,
				Path:    "/upload",
				Handler: UploadHandler(serverCtx),
			},
		},
	)
}
