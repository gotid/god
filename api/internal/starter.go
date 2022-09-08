package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gotid/god/lib/logx"

	"github.com/gotid/god/lib/proc"
)

// StartOption 自定义 http.Server 的方法。
type StartOption func(s *http.Server)

// StartHttp 启动一个 http 服务器。
func StartHttp(host string, port int, handler http.Handler, opts ...StartOption) error {
	return start(host, port, handler, func(server *http.Server) error {
		return server.ListenAndServe()
	}, opts...)
}

// StartHttps 启动一个 https server。
func StartHttps(host string, port int, certFile, keyFile string, handler http.Handler,
	opts ...StartOption) error {
	return start(host, port, handler, func(server *http.Server) error {
		// 证书文件和秘钥文件在 buildHttpsServer 中设置
		return server.ListenAndServeTLS(certFile, keyFile)
	}, opts...)
}

func start(host string, port int, handler http.Handler, run func(server *http.Server) error,
	opts ...StartOption) (err error) {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: handler,
	}
	for _, opt := range opts {
		opt(server)
	}

	waitForCalled := proc.AddWrapUpListener(func() {
		if e := server.Shutdown(context.Background()); err != nil {
			logx.Error(e)
		}
	})
	defer func() {
		if err == http.ErrServerClosed {
			waitForCalled()
		}
	}()

	return run(server)
}
