package handler

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"git.zc0901.com/go/god/api/httpx"
	"git.zc0901.com/go/god/api/internal"
	"git.zc0901.com/go/god/lib/iox"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/timex"
	"git.zc0901.com/go/god/lib/utils"
)

const (
	limitBodyBytes = 1024
	slowThreshold  = 500 * time.Millisecond // 慢日志阈值
)

// LogHandler 返回一个简要记录请求和响应的日志中间件。
func LogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := utils.NewElapsedTimer()
		logs := new(internal.LogCollector)
		lrw := loggedResponseWriter{
			w:    w,
			r:    r,
			code: http.StatusOK,
		}

		var dup io.ReadCloser
		r.Body, dup = iox.DupReadCloser(r.Body)
		next.ServeHTTP(&lrw, r.WithContext(context.WithValue(r.Context(), internal.LogContext, logs)))
		r.Body = dup
		logBrief(r, lrw.code, timer, logs)
	})
}

// DetailedLogHandler API 详细日志记录中间件
func DetailedLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := utils.NewElapsedTimer()
		var buf bytes.Buffer
		lrw := newDetailLoggedResponseWriter(&loggedResponseWriter{
			w:    w,
			r:    r,
			code: http.StatusOK,
		}, &buf)

		var dup io.ReadCloser
		r.Body, dup = iox.DupReadCloser(r.Body)
		logs := new(internal.LogCollector)
		next.ServeHTTP(lrw, r.WithContext(context.WithValue(r.Context(), internal.LogContext, logs)))
		r.Body = dup
		logDetails(r, lrw, timer, logs)
	})
}

type detailLoggedResponseWriter struct {
	writer *loggedResponseWriter
	buf    *bytes.Buffer
}

func newDetailLoggedResponseWriter(writer *loggedResponseWriter, buf *bytes.Buffer) *detailLoggedResponseWriter {
	return &detailLoggedResponseWriter{
		writer: writer,
		buf:    buf,
	}
}

func (w *detailLoggedResponseWriter) Header() http.Header {
	return w.writer.Header()
}

// Hijack implements the http.Hijacker interface.
// This expands the Response to fulfill http.Hijacker if the underlying http.ResponseWriter supports it.
func (w *detailLoggedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacked, ok := w.writer.w.(http.Hijacker); ok {
		return hijacked.Hijack()
	}

	return nil, nil, errors.New("server doesn't support hijacking")
}

func (w *detailLoggedResponseWriter) Write(bs []byte) (int, error) {
	w.buf.Write(bs)
	return w.writer.Write(bs)
}

func (w *detailLoggedResponseWriter) WriteHeader(code int) {
	w.writer.WriteHeader(code)
}

func dumpRequest(r *http.Request) string {
	reqContent, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err.Error()
	} else {
		return string(reqContent)
	}
}

// 记录摘要。
func logBrief(r *http.Request, code int, timer *utils.ElapsedTimer, logs *internal.LogCollector) {
	var buf bytes.Buffer
	duration := timer.Duration()
	logger := logx.WithContext(r.Context())
	buf.WriteString(fmt.Sprintf("[HTTP] %s - %d - %s - %s - %s - %s",
		r.Method, code, r.RequestURI, httpx.GetRemoteAddr(r), r.UserAgent(), timex.MillisecondDuration(duration)))
	if duration > slowThreshold {
		logger.Slowf("[HTTP] %s - %d - %s - %s - %s - 慢调用(%s)",
			r.Method, code, r.RequestURI, httpx.GetRemoteAddr(r), r.UserAgent(), timex.MillisecondDuration(duration))
	}

	ok := isOkResponse(code)
	if !ok {
		fullReq := dumpRequest(r)
		limitReader := io.LimitReader(strings.NewReader(fullReq), limitBodyBytes)
		body, err := ioutil.ReadAll(limitReader)
		if err != nil {
			buf.WriteString(fmt.Sprintf("\n%s", fullReq))
		} else {
			buf.WriteString(fmt.Sprintf("\n%s", string(body)))
		}
	}

	body := logs.Flush()
	if len(body) > 0 {
		buf.WriteString(fmt.Sprintf("\n%s", body))
	}

	if ok {
		logger.Info(buf.String())
	} else {
		logger.Error(buf.String())
	}
}

// 记录明细。
func logDetails(r *http.Request, response *detailLoggedResponseWriter, timer *utils.ElapsedTimer,
	logs *internal.LogCollector) {
	var buf bytes.Buffer
	duration := timer.Duration()
	logger := logx.WithContext(r.Context())
	buf.WriteString(fmt.Sprintf("[HTTP] %s - %d - %s - %s\n=> %s\n",
		r.Method, response.writer.code, r.RemoteAddr, timex.MillisecondDuration(duration), dumpRequest(r)))
	if duration > slowThreshold {
		logger.Slowf("[HTTP] %s - %d - %s - 慢调用(%s)\n=> %s\n",
			r.Method, response.writer.code, r.RemoteAddr, timex.MillisecondDuration(duration), dumpRequest(r))
	}

	body := logs.Flush()
	if len(body) > 0 {
		buf.WriteString(fmt.Sprintf("%s\n", body))
	}

	respBuf := response.buf.Bytes()
	if len(respBuf) > 0 {
		buf.WriteString(fmt.Sprintf("<= %s", respBuf))
	}

	logger.Info(buf.String())
}

func isOkResponse(code int) bool {
	// 非内部服务器错误
	return code < http.StatusInternalServerError
}

// 带日志的响应输出器。
type loggedResponseWriter struct {
	w    http.ResponseWriter
	r    *http.Request
	code int
}

func (w *loggedResponseWriter) Flush() {
	if flusher, ok := w.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *loggedResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *loggedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacked, ok := w.w.(http.Hijacker); ok {
		return hijacked.Hijack()
	}

	return nil, nil, errors.New("服务器不支持 Hijacker")
}

func (w *loggedResponseWriter) Write(bytes []byte) (int, error) {
	return w.w.Write(bytes)
}

func (w *loggedResponseWriter) WriteHeader(code int) {
	w.w.WriteHeader(code)
	w.code = code
}
