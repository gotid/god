package main

import (
	"net/http"
	"runtime"
	"time"

	"git.zc0901.com/go/god/api"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/service"
	"git.zc0901.com/go/god/lib/stat"
	"git.zc0901.com/go/god/lib/syncx"
)

func main() {
	logx.Disable()
	stat.SetReporter(nil)

	server := api.MustNewServer(api.ServerConf{
		ServiceConf: service.ServiceConf{
			Name: "api breaker",
			Log:  logx.LogConf{Mode: "console"},
		},
		Host:     "0.0.0.0",
		Port:     8080,
		MaxConns: 1000,
		Timeout:  3000,
	})
	defer server.Stop()

	limiter := syncx.NewLimit(10)

	server.AddRoute(api.Route{
		Method: http.MethodGet,
		Path:   "/heavy",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if limiter.TryBorrow() {
				defer limiter.Return()
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

	server.Start()
}
