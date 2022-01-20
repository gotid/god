package main

import (
	"flag"
	"net/http"

	"git.zc0901.com/go/god/example/wechat/pkg"

	"git.zc0901.com/go/god/api"
	"git.zc0901.com/go/god/lib/conf"
	"git.zc0901.com/go/god/lib/logx"
)

var (
	configFile = flag.String("f", "config.yaml", "API 配置文件")
	handlers   *pkg.OpenHandlers
)

func init() {
	handlers = pkg.NewOpenHandlers()
}

func main() {
	flag.Parse()
	var apiConf api.ServerConf
	conf.MustLoad(*configFile, &apiConf)

	server := api.MustNewServer(apiConf)
	defer server.Stop()

	server.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/home",
		Handler: homeHandler,
	})

	// 授权事件接收配置
	server.AddRoute(api.Route{
		Method:  http.MethodPost,
		Path:    "/oplatform/wxbef357be217c23c5/notify",
		Handler: handlers.Notify,
	})

	// 获取 component_verify_ticket 后，查看平台令牌
	server.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/oplatform/wxbef357be217c23c5/accesstoken",
		Handler: handlers.AccessToken,
	})

	// 生成PC版/移动版平台授权码 ?isMobile=1
	server.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/oplatform/wxbef357be217c23c5/auth",
		Handler: handlers.Auth,
	})

	// 授权方授权后跳转网址
	server.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/oplatform/wxbef357be217c23c5/redirect",
		Handler: handlers.Redirect,
	})

	// 根据微信重定向带过来的授权码查询授权信息
	server.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/oplatform/wxbef357be217c23c5/queryauth",
		Handler: handlers.QueryAuth,
	})

	logx.Infof("微信响应 API NewServer 已启动 —— %v:%v", apiConf.Host, apiConf.Port)
	server.Start()
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	w.Write([]byte("hello world"))
}
