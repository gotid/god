package main

import (
	"flag"
	"github.com/gotid/god/api"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/lib/logx"
	"net/http"
)

var configFile = flag.String("f", "config.yaml", "配置文件")

type Request struct {
	User string `form:"user"`
}

func first(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Middleware", "first")
		next(w, r)
	}
}

func second(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Middleware", "second")
		next(w, r)
	}
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := httpx.Parse(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	httpx.OkJson(w, "welcome, "+req.User)
}

func main() {
	flag.Parse()
	logx.DisableStat()

	var c api.Config
	conf.MustLoad(*configFile, &c)
	server := api.MustNewServer(c, api.WithCors("http://localhost:8100", "https://zhuke.com"))
	defer server.Stop()

	server.Use(first)
	server.Use(second)

	server.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/hello",
		Handler: handleHello,
	})

	server.Start()
}
