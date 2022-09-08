package response

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type WithCodeResponseWriter struct {
	Writer http.ResponseWriter
	Code   int
}

// Header 返回 http 头。
func (w *WithCodeResponseWriter) Header() http.Header {
	return w.Writer.Header()
}

// Write 写入字节到 http 作为响应。
func (w *WithCodeResponseWriter) Write(bytes []byte) (int, error) {
	return w.Writer.Write(bytes)
}

// WriteHeader 一次性写入响应状态码，且不终止响应。
func (w *WithCodeResponseWriter) WriteHeader(statusCode int) {
	w.Writer.WriteHeader(statusCode)
	w.Code = statusCode
}

// Flush 刷新响应编写器。
func (w *WithCodeResponseWriter) Flush() {
	if flusher, ok := w.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack 实现了 http.Hijacker 接口。
// 如果底层 http.ResponseWriter 支持它，这将扩展响应以实现 http.Hijacker。
func (w *WithCodeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacked, ok := w.Writer.(http.Hijacker); ok {
		return hijacked.Hijack()
	}

	return nil, nil, errors.New("服务器不支持 hijack")
}
