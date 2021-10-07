// Code generated by god. DO NOT EDIT.
package handler

import (
	"net/http"

	"git.zc0901.com/go/god/example/test/cmd/api/internal/svc"

	"git.zc0901.com/go/god/api"
)

func RegisterHandlers(engine *api.Server, serverCtx *svc.ServiceContext) {
	engine.AddRoutes(
		[]api.Route{
			{
				Method:  http.MethodGet,
				Path:    "/ping/:name",
				Handler: PingHandler(serverCtx),
			},
		},
	)
}
