package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/api/internal"
	"io"
	"net/http"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	statusClientClosedRequest = 499
	reason                    = "Request Timeout"
	headerUpgrade             = "Upgrade"
	valueWebsocket            = "websocket"
)

// TimeoutHandler 返回一个超时控制中间件。
// 如果客户端关闭请求，状态码将记录为 499。
// 注意：计时在服务端取消，也会被记录为 499.
func TimeoutHandler(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if duration > 0 {
			return &timeoutHandler{
				handler: next,
				dt:      duration,
			}
		}

		return next
	}
}

// timeoutHandler 是控制请求超时的处理器。
// 之所以我们自己实现，原因是标准库会把客户端关闭请求当做服务不可用（503），
// 而我们遵照 nginx 将该错误定义为 499。
type timeoutHandler struct {
	handler http.Handler
	dt      time.Duration
}

func (h *timeoutHandler) errorBody() string {
	return reason
}

func (h *timeoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(headerUpgrade) == valueWebsocket {
		h.handler.ServeHTTP(w, r)
		return
	}

	ctx, cancelCtx := context.WithTimeout(r.Context(), h.dt)
	defer cancelCtx()

	r = r.WithContext(ctx)
	done := make(chan struct{})
	tw := &timeoutWriter{
		w:   w,
		h:   make(http.Header),
		req: r,
	}

	panicChan := make(chan interface{}, 1)

	go func() {
		defer func() {
			if p := recover(); p != nil {
				panicChan <- p
			}
		}()
		h.handler.ServeHTTP(tw, r)
		close(done)
	}()

	select {
	case p := <-panicChan:
		panic(p)
	case <-done:
		tw.mu.Lock()
		defer tw.mu.Unlock()
		dst := w.Header()
		for k, vv := range tw.h {
			dst[k] = vv
		}
		if !tw.wroteHeader {
			tw.code = http.StatusOK
		}
		w.WriteHeader(tw.code)
		w.Write(tw.wbuf.Bytes())
	case <-ctx.Done():
		tw.mu.Lock()
		defer tw.mu.Unlock()
		// TimoutHandler 之前没有任何用户定义的中间件，
		// 所以我们可以保证，业务代码中的取消不会出现于此。
		httpx.Error(w, ctx.Err(), func(w http.ResponseWriter, err error) {
			if errors.Is(err, context.Canceled) {
				w.WriteHeader(statusClientClosedRequest)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			io.WriteString(w, h.errorBody())
		})
		tw.timedOut = true
	}
}

type timeoutWriter struct {
	w    http.ResponseWriter
	h    http.Header
	wbuf bytes.Buffer
	req  *http.Request

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
	code        int
}

var _ http.Pusher = (*timeoutWriter)(nil)

// Header returns the underline temporary http.Header.
func (tw *timeoutWriter) Header() http.Header { return tw.h }

// Push implements the Pusher interface.
func (tw *timeoutWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := tw.w.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// Write writes the data to the connection as part of an HTTP reply.
// Timeout and multiple header written are guarded.
func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut {
		return 0, http.ErrHandlerTimeout
	}

	if !tw.wroteHeader {
		tw.writeHeaderLocked(http.StatusOK)
	}
	return tw.wbuf.Write(p)
}

func (tw *timeoutWriter) writeHeaderLocked(code int) {
	checkWriteHeaderCode(code)

	switch {
	case tw.timedOut:
		return
	case tw.wroteHeader:
		if tw.req != nil {
			caller := relevantCaller()
			internal.Errorf(tw.req, "http: superfluous response.WriteHeader call from %s (%s:%d)",
				caller.Function, path.Base(caller.File), caller.Line)
		}
	default:
		tw.wroteHeader = true
		tw.code = code
	}
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.writeHeaderLocked(code)
}

func checkWriteHeaderCode(code int) {
	if code < 100 || code > 599 {
		panic(fmt.Sprintf("无效的状态码 %v", code))
	}
}

// relevantCaller 在调用堆栈中搜索net/http之外的第一个函数。
// 此函数的目的是提供更有用的错误消息。
func relevantCaller() runtime.Frame {
	pc := make([]uintptr, 16)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
	for {
		frame, more := frames.Next()
		if !strings.HasPrefix(frame.Function, "net/http.") {
			return frame
		}
		if !more {
			break
		}
	}
	return frame
}
