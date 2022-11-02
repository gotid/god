package response

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// WithCodeResponseWriter 是一个带状态码的响应编写器。
type WithCodeResponseWriter struct {
	Writer http.ResponseWriter
	Code   int
}

func (w *WithCodeResponseWriter) Header() http.Header {
	return w.Writer.Header()
}

func (w *WithCodeResponseWriter) Write(bytes []byte) (int, error) {
	return w.Writer.Write(bytes)
}

func (w *WithCodeResponseWriter) WriteHeader(statusCode int) {
	w.Writer.WriteHeader(statusCode)
	w.Code = statusCode
}

// Hijack 实现 http.Hijacker 接口。
// 如果底层 http.ResponseWriter 支持的话，该方法将扩展响应以满足 http.Hijacker。
func (w *WithCodeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacked, ok := w.Writer.(http.Hijacker); ok {
		return hijacked.Hijack()
	}

	return nil, nil, errors.New("服务器不支持 hijacking 劫持")
}

// Flush 发送缓冲数据到客户端。
func (w *WithCodeResponseWriter) Flush() {
	if flusher, ok := w.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}
