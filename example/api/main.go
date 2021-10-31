package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"git.zc0901.com/go/god/lib/g"

	"git.zc0901.com/go/god/lib/syncx"

	"git.zc0901.com/go/god/api"
	"git.zc0901.com/go/god/api/httpx"
	"git.zc0901.com/go/god/lib/conf"
)

var configFile = flag.String("f", "config.yaml", "API 配置文件")

func main() {
	// 读取配置文件
	flag.Parse()
	var c api.Conf
	conf.MustLoad(*configFile, &c)

	// 新建 API 服务器
	server := api.MustNewServer(c, api.WithNotAllowedHandler(api.CorsHandler()))
	defer server.Stop()

	// 增加路由
	server.AddRoute(api.Route{
		Method: http.MethodGet,
		Path:   "/api",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			httpx.OkJson(w, map[string]string{
				"data": fmt.Sprintf("hello, world!-%d", time.Now().UnixMilli()),
			})
		},
	})

	// 模拟并发限制
	limiter := syncx.NewLimit(2)
	server.AddRoute(api.Route{
		Method: http.MethodGet,
		Path:   "/api/ping",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if limiter.TryBorrow() {
				defer limiter.Return()
				runtime.LockOSThread()
				defer runtime.UnlockOSThread()
				begin := time.Now()
				for {
					if time.Now().Sub(begin) > 9*time.Millisecond {
						break
					}
				}
				httpx.OkJson(w, g.Map{"data": "pong"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		},
	})

	// 启动 API 服务器
	server.Start()
}
