package main

import (
	"fmt"
	"net/http"

	"git.zc0901.com/go/god/lib/wechat/msg"

	cacheRedis "git.zc0901.com/go/god/lib/store/cache"
	"git.zc0901.com/go/god/lib/store/kv"
	"git.zc0901.com/go/god/lib/store/redis"
	"git.zc0901.com/go/god/lib/wechat"
	"git.zc0901.com/go/god/lib/wechat/cache"
	"git.zc0901.com/go/god/lib/wechat/context"
)

var ctx *context.Context

func init() {
	store := kv.NewStore([]cacheRedis.Conf{
		{
			Conf: redis.Conf{
				Host:     "vps:6382",
				Password: "4a5d4787a82c660ee18719f51ff40d9a669a4958",
				Mode:     redis.StandaloneMode,
			},
			Weight: 100,
		},
	})

	ctx = &context.Context{
		AppID:          "wxbef357be217c23c5",
		AppSecret:      "403d127716317ea23c8db1a1107b14fc",
		Token:          "imola1999zhuke2012dhome2020",
		EncodingAESKey: "imola1999azhuke2012adhome2020a18611914900aa",
		Cache:          cache.NewRedis(store),
	}
}

func main() {
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Printf("启动服务器错误，错误=%v", err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	wc := wechat.New(ctx)
	server := wc.NewServer(w, r)

	// 设置常规消息钩子
	server.SetMsgHandler(func(m msg.Msg) *msg.Response {
		return &msg.Response{
			Scene: msg.ResponseSceneKefu,
			Type:  msg.ResponseTypeXML,
			Msg:   "hello world",
		}
	})

	// 处理请求、构建响应
	err := server.Serve()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 发送响应
	server.Send()
}
