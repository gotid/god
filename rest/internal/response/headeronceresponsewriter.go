package response

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// HeaderOnceResponseWriter 是一个响应编写器 http.ResponseWriter。
// 其特点是：只有第一此 WriterHeader 写入的标头会生效。
type HeaderOnceResponseWriter struct {
	w           http.ResponseWriter
	wroteHeader bool
}

// NewHeaderOnceResponseWriter 返回一个 HeaderOnceResponseWriter。
func NewHeaderOnceResponseWriter(w http.ResponseWriter) http.ResponseWriter {
	return &HeaderOnceResponseWriter{w: w}
}

func (w *HeaderOnceResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *HeaderOnceResponseWriter) Write(bytes []byte) (int, error) {
	return w.w.Write(bytes)
}

func (w *HeaderOnceResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.w.WriteHeader(statusCode)
	w.wroteHeader = true
}

// Hijack 实现 http.Hijacker 接口。
// 如果底层 http.ResponseWriter 支持的话，该方法将扩展响应以满足 http.Hijacker。
func (w *HeaderOnceResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacked, ok := w.w.(http.Hijacker); ok {
		return hijacked.Hijack()
	}

	return nil, nil, errors.New("服务器不支持 hijacking 劫持")
}

// Flush 发送缓冲数据到客户端。
func (w *HeaderOnceResponseWriter) Flush() {
	if flusher, ok := w.w.(http.Flusher); ok {
		flusher.Flush()
	}
}
