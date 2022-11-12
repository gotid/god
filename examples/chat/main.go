package main

import (
	"flag"
	"github.com/gotid/god/api"
	"github.com/gotid/god/examples/chat/internal"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/service"
	"net/http"
)

var (
	port    = flag.Int("port", 3333, "监听端口")
	timeout = flag.Int64("timeout", 0, "超时毫秒数")
	cpu     = flag.Int64("cpu", 500, "cpu 自动降载阈值")
)

func main() {
	flag.Parse()

	logx.Disable()
	engine := api.MustNewServer(api.Config{
		Config: service.Config{
			Log: logx.Config{
				Mode: "console",
			},
		},
		Host:         "localhost",
		Port:         *port,
		Timeout:      *timeout,
		CpuThreshold: *cpu,
	})
	defer engine.Stop()

	hub := internal.NewHub()
	go hub.Run()

	engine.AddRoute(api.Route{
		Method: http.MethodGet,
		Path:   "/",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.Error(w, "资源未找到", http.StatusNotFound)
				return
			}
			if r.Method != "GET" {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
				return
			}

			http.ServeFile(w, r, "/Users/zs/Github/gotid/god/examples/chat/home.html")
		},
	})

	engine.AddRoute(api.Route{
		Method: http.MethodGet,
		Path:   "/ws",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			internal.ServeWs(hub, w, r)
		},
	})

	engine.Start()
}
