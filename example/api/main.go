package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gotid/god/lib/g"

	"github.com/gotid/god/lib/syncx"

	"github.com/gotid/god/api"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/lib/conf"
)

var configFile = flag.String("f", "config.yaml", "API 配置文件")

type ApiMessage struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
	Msg  string      `json:"message,omitempty"`
}

func ApiErrorHandler(err error) (int, interface{}) {
	return http.StatusOK, ApiMessage{
		Code: 0,
		Msg:  err.Error(),
	}
}

func ApiOKHandler(data interface{}) interface{} {
	return ApiMessage{
		Code: 0,
		Data: data,
	}
}

func main() {
	// 读取配置文件
	flag.Parse()
	var c api.ServerConf
	conf.MustLoad(*configFile, &c)

	// 新建 API 服务器
	server := api.MustNewServer(c,
		api.WithCors(),
		api.WithNotFoundHandler(NewNotFound()),
	)
	defer server.Stop()

	httpx.SetErrorHandler(ApiErrorHandler)
	httpx.SetOkJsonHandler(ApiOKHandler)

	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if p := recover(); p != nil {
				httpx.WriteJson(w, http.StatusOK, fmt.Errorf("❎出错啦🌶, %v", p))
			} else {
				next.ServeHTTP(w, r)
			}
		}
	})

	// 增加路由
	server.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/api",
		Handler: apiHandler,
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

func apiHandler(w http.ResponseWriter, r *http.Request) {
	a, b := 1, 0
	fmt.Println(a / b)
	httpx.OkJson(w, map[string]string{
		"data": fmt.Sprintf("hello, world!-%d", time.Now().UnixMilli()),
	})
}

type NotFound struct{}

func NewNotFound() *NotFound {
	return &NotFound{}
}

func (n *NotFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	msg := ApiMessage{
		Code: http.StatusNotFound,
		Msg:  "接口不存在",
	}
	httpx.WriteJson(w, http.StatusOK, msg)
}

func UnauthorizedCallback() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		msg := ApiMessage{
			Code: http.StatusUnauthorized,
			Msg:  "请登录",
		}
		httpx.WriteJson(w, http.StatusOK, msg)
	}
}
