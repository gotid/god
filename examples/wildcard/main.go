package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/api"
	"github.com/gotid/god/api/httpx"
	"net/http"
	"strings"
)

var port = flag.Int("port", 3333, "监听端口")

type Request struct {
	User string `form:"user,options=a|b"`
}

func handle(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := httpx.Parse(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	httpx.OkJson(w, "hello, "+req.User)
}

func main() {
	flag.Parse()

	svr := api.MustNewServer(api.Config{
		Host: "localhost",
		Port: *port,
	}, api.WithNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/any/") {
			fmt.Fprintf(w, "wildcard: %s", r.URL.Path)
		} else {
			http.NotFound(w, r)
		}
	})))
	defer svr.Stop()

	svr.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/api/users",
		Handler: handle,
	})

	svr.Start()
}
