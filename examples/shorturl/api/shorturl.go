package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gotid/god/api"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/lib/conf"

	"github.com/gotid/god/examples/shorturl/api/internal/config"
	"github.com/gotid/god/examples/shorturl/api/internal/handler"
	"github.com/gotid/god/examples/shorturl/api/internal/svc"
)

var configFile = flag.String("f", "etc/shorturl-api.yaml", "配置文件")

// Message 返回的结构体，json格式的body
type Message struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func main() {
	flag.Parse()

	// 设置错误处理函数
	httpx.SetErrorHandler(func(err error) (int, any) {
		return http.StatusConflict, Message{
			Code: -1,
			Msg:  err.Error(),
		}
	})

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := api.MustNewServer(c.Config)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("启动 api 服务器 %s:%d...\n", c.Host, c.Port)
	server.Start()
}
