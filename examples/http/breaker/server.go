package main

import (
	"flag"
	"github.com/gotid/god/api"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/syncx"
	"net/http"
	"runtime"
	"time"
)

var port = flag.Int("p", 8888, "端口号")

func main() {
	flag.Parse()
	logx.Disable()
	stat.SetReporter(nil)

	server := api.MustNewServer(api.Config{
		Config: service.Config{
			Name: "breaker",
			Log: logx.Config{
				Mode:     "console",
				Encoding: "plain",
			},
		},
		Host:     "0.0.0.0",
		Port:     *port,
		MaxConns: 1000,
		Timeout:  3000,
	})

	latch := syncx.NewLimit(10)
	server.AddRoute(api.Route{
		Method: http.MethodGet,
		Path:   "/heavy",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if latch.TryBorrow() {
				defer latch.Return()
				runtime.LockOSThread()
				defer runtime.UnlockOSThread()
				begin := time.Now()
				for {
					if time.Now().Sub(begin) > time.Millisecond*50 {
						break
					}
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		},
	})
	server.AddRoute(api.Route{
		Method: http.MethodGet,
		Path:   "/good",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	})
	defer server.Stop()
	server.Start()
}
