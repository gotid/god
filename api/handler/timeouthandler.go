package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gotid/god/api/httpx"

	"github.com/gotid/god/api/internal"
)

const (
	reason                    = "请求超时"
	statusClientClosedRequest = 499
)

// TimeoutHandler 返回指定超时时长的中间件。
// 如果客户端或服务端关闭请求，状态码都将被记录为 499。
func TimeoutHandler(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if duration > 0 {
			return &timeoutHandler{
				handler:  next,
				duration: duration,
			}
		}

		return next
	}
}

// 是控制请求超时的处理器。
// 之所以自己实现是因为标准实现把客户端关闭请求 ClientClosedRequest 当成了 http.StatusServiceUnavailable。
// 我们参考 nginx 定义将此响应码改写为 499。
type timeoutHandler struct {
	handler  http.Handler
	duration time.Duration
}

func (h *timeoutHandler) errorBody() string {
	return reason
}

func (h *timeoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancelFunc := context.WithTimeout(r.Context(), h.duration)
	defer cancelFunc()

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
		w.Write(tw.buf.Bytes())
	case <-ctx.Done():
		tw.mu.Lock()
		defer tw.mu.Unlock()
		// 因为在 TimeoutHandler 之前没有任何用户定义的中间件，
		// 所以我们可以保证业务相关的取消代码不会出现在这里。
		httpx.Error(w, ctx.Err(), func(w http.ResponseWriter, err error) {
			if errors.Is(ctx.Err(), context.Canceled) {
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
	w   http.ResponseWriter
	h   http.Header
	req *http.Request
	buf bytes.Buffer

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
	code        int
}

var _ http.Pusher = (*timeoutWriter)(nil)

// Push 实现 Pusher 接口。
func (tw *timeoutWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := tw.w.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (tw *timeoutWriter) Header() http.Header { return tw.h }

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut {
		return 0, http.ErrHandlerTimeout
	}

	if !tw.wroteHeader {
		tw.writeHeaderLocked(http.StatusOK)
	}
	return tw.buf.Write(p)
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
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
}

// relevantCaller searches the call stack for the first function outside of net/http.
// The purpose of this function is to provide more helpful error messages.
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
