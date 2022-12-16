package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gotid/god/internal/health"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/proc"
)

const probeNamePrefix = "api"

// StartOption 自定义 http.Server 的方法。
type StartOption func(svr *http.Server)

// StartHttp 启动一个 http 服务器。
func StartHttp(host string, port int, handler http.Handler, opts ...StartOption) error {
	return start(host, port, handler, func(svr *http.Server) error {
		return svr.ListenAndServe()
	}, opts...)
}

// StartHttps 启动一个 https 服务器。
func StartHttps(host string, port int, certFile, keyFile string, handler http.Handler, opts ...StartOption) error {
	return start(host, port, handler, func(svr *http.Server) error {
		// certFile 证书文件和 keyFile 秘钥文件在 buildHttpsServer 中设置
		return svr.ListenAndServeTLS(certFile, keyFile)
	}, opts...)
}

// 创建一个带有自定义选项 http.Server 实例，并记录关闭结果。
func start(host string, port int, handler http.Handler, run func(svr *http.Server) error, opts ...StartOption) (err error) {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: handler,
	}
	for _, opt := range opts {
		opt(server)
	}

	healthManager := health.NewHealthManager(fmt.Sprintf("%s-%s:%d", probeNamePrefix, host, port))

	waitForCalled := proc.AddWrapUpListener(func() {
		healthManager.MarkNotReady()
		if e := server.Shutdown(context.Background()); e != nil {
			logx.Error(e)
		}
	})

	defer func() {
		if err == http.ErrServerClosed {
			waitForCalled()
		}
	}()

	healthManager.MarkReady()
	health.AddProbe(healthManager)

	return run(server)
}
