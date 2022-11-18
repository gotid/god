package main

import (
	"github.com/gotid/god/api"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/examples/tracing/remote/portal"
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/service"
	"net/http"

	"github.com/gotid/god/rpc"
)

var client rpc.Client

func handle(w http.ResponseWriter, r *http.Request) {
	conn := client.Conn()
	greet := portal.NewPortalClient(conn)
	resp, err := greet.Portal(r.Context(), &portal.PortalRequest{
		Name: "richard",
	})
	if err != nil {
		httpx.WriteJson(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		httpx.OkJson(w, resp.Response)
	}
}

func main() {
	client = rpc.MustNewClient(rpc.ClientConfig{
		Etcd: discov.EtcdConfig{
			Hosts: []string{
				"localhost:2379",
			},
			Key: "portal",
		},
	})
	engine := api.MustNewServer(api.Config{
		Config: service.Config{
			Name: "edge-api",
		},
		Host: "0.0.0.0",
		Port: 3456,
	})
	defer engine.Stop()

	engine.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/",
		Handler: handle,
	})
	engine.Start()
}
