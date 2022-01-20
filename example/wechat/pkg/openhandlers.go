package pkg

import (
	"fmt"
	"net/http"

	"git.zc0901.com/go/god/lib/g"

	"git.zc0901.com/go/god/lib/gconv"

	"git.zc0901.com/go/god/lib/store/kv"
	"git.zc0901.com/go/god/lib/store/redis"

	"git.zc0901.com/go/god/api/httpx"

	"git.zc0901.com/go/god/lib/logx"
	cacheRedis "git.zc0901.com/go/god/lib/store/cache"
	"git.zc0901.com/go/god/lib/wechat/cache"
	"git.zc0901.com/go/god/lib/wechat/msg"
	"git.zc0901.com/go/god/lib/wechat/openplatform"
	"git.zc0901.com/go/god/lib/wechat/openplatform/config"
)

type OpenHandlers struct {
	open *openplatform.OpenPlatform
}

func NewOpenHandlers() *OpenHandlers {
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

	o := openplatform.New(&config.Config{
		AppID:          "wxbef357be217c23c5",
		AppSecret:      "403d127716317ea23c8db1a1107b14fc",
		Token:          "imola1999zhuke2012dhome2020",
		EncodingAESKey: "imola1999azhuke2012adhome2020a18611914900aa",
		Cache:          cache.NewRedis(store),
	})
	return &OpenHandlers{o}
}

func (h *OpenHandlers) Notify(w http.ResponseWriter, r *http.Request) {
	logx.Debugf("请求方法: %v", r.Method)
	server := openplatform.GetServer(h.open.Context, w, r)

	// 设置消息响应处理器
	server.SetMessageHandler(func(m *msg.Msg) (*msg.Response, error) {
		if m.InfoType == "component_verify_ticket" {
			accessToken, err := h.open.SetAccessToken(m.ComponentVerifyTicket)
			if err != nil {
				return nil, err
			}
			logx.Debug("平台访问令牌", accessToken)
		} else {
			fmt.Println("消息钩子", m)
		}
		return nil, nil
	})

	// 处理微信的请求信息
	err := server.Serve()
	if err != nil {
		logx.Error("微信响应服务器错误：", err)
		return
	}

	// 发送回复的消息
	err = server.Send()
	if err != nil {
		logx.Errorf("响应微信错误，err=%+v", err)
		return
	}
}

func (h *OpenHandlers) AccessToken(w http.ResponseWriter, r *http.Request) {
	accessToken, err := h.open.AccessToken()
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.OkJson(w, map[string]string{
		"token": accessToken,
	})
}

func (h *OpenHandlers) Auth(w http.ResponseWriter, r *http.Request) {
	isMobile := false
	if vs, ok := r.URL.Query()["isMobile"]; ok && len(vs) == 1 {
		isMobile = gconv.Bool(vs[0])
	}

	apiHost := "http://zs.ngrok.zc0901.com"
	redirect := fmt.Sprintf("%s/oplatform/%s/redirect", apiHost, h.open.AppID)

	var authUrl string
	var err error

	if isMobile {
		authUrl, err = h.open.MobileAuthURL(redirect, 2, "")
	} else {
		authUrl, err = h.open.PcAuthURL(redirect, 2, "")
	}
	if err != nil {
		w.WriteHeader(200)
		w.Write([]byte("系统异常: " + err.Error()))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf("<script>location.href=\"%s\"</script>", authUrl)))
}

func (h *OpenHandlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var authCode, expiresIn string

	// 授权码
	if vs := r.URL.Query()["auth_code"]; len(vs) == 1 {
		authCode = vs[0]
	}

	// 过期时间
	if vs := r.URL.Query()["expires_in"]; len(vs) == 1 {
		expiresIn = vs[0]
	}

	http.Redirect(w, r, "/home?auth_code="+authCode+"&expires_in="+expiresIn, 302)
}

func (h *OpenHandlers) QueryAuth(w http.ResponseWriter, r *http.Request) {
	var authCode string

	// 授权码
	if vs := r.URL.Query()["auth_code"]; len(vs) == 1 {
		authCode = vs[0]
	}

	authInfo, err := h.open.QueryAuth(authCode)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.OkJson(w, g.Map{
		"authInfo": authInfo,
	})
}
