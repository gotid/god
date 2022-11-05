package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/api"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/examples/http/signature/internal"
	"github.com/gotid/god/lib/fs"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/service"
	"io"
	"net/http"
	"os"
	"time"
)

type Request struct {
	User string `form:"user,optional"`
}

func main() {
	flag.Parse()
	priFile, err := fs.TempFilenameWithText(string(internal.PriKey))
	defer os.Remove(priFile)
	if err != nil {
		panic(err)
	}

	c := api.Config{
		Config: service.Config{
			Log: logx.Config{
				Mode: "console",
			},
		},
		Verbose: true,
		Port:    3333,
		Signature: api.SignatureConfig{
			Strict: true,
			Expire: 10 * time.Minute,
			PrivateKeys: []api.PrivateKeyConfig{
				{
					Fingerprint: internal.Fingerprint,
					KeyFile:     priFile,
				},
			},
		},
	}

	engine := api.MustNewServer(c)
	defer engine.Stop()

	engine.AddRoute(api.Route{
		Method:  http.MethodPost,
		Path:    "/a/b",
		Handler: handler,
	}, api.WithSignature(c.Signature))

	fmt.Println("启动服务器...")
	engine.Start()
}

func handler(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := httpx.Parse(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	io.Copy(w, r.Body)
	//w.Write([]byte("hello world!"))
}
