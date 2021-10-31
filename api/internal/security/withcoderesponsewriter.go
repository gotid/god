package security

import (
	"bufio"
	"net"
	"net/http"
)

// WithCodeResponseWriter 带响应码、延迟输出的 http.ResponseWriter 助手结构体。
type WithCodeResponseWriter struct {
	Writer http.ResponseWriter
	Code   int
}

// Flush 刷新响应输出器。
func (w *WithCodeResponseWriter) Flush() {
	if flusher, ok := w.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Header 返回 HTTP 标头。
func (w *WithCodeResponseWriter) Header() http.Header {
	return w.Writer.Header()
}

// Hijack 实现 http.Hijacker 接口。
// 如果底层 http.ResponseWriter 支持，此举将扩展实现 http.Hijacker。
func (w *WithCodeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.Writer.(http.Hijacker).Hijack()
}

// Write 将字节数组写入响应流。
func (w *WithCodeResponseWriter) Write(bytes []byte) (int, error) {
	return w.Writer.Write(bytes)
}

// WriteHeader 写入状态码但并不封装输出器。
func (w *WithCodeResponseWriter) WriteHeader(code int) {
	w.Writer.WriteHeader(code)
	w.Code = code
}
