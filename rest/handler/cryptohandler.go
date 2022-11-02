package handler

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/gotid/god/lib/codec"
	"github.com/gotid/god/lib/logx"
	"io"
	"net"
	"net/http"
)

const maxBytes = 1 << 20 // 1 MiB

var errContentLengthExceeded = errors.New("内容长度越界")

// CryptoHandler 返回一个给定加密键的内容自动加解密中间件。
func CryptoHandler(key []byte) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cw := newCryptoResponseWriter(w)
			defer cw.flush(key)

			if r.ContentLength <= 0 {
				next.ServeHTTP(cw, r)
				return
			}

			if err := decryptBody(key, r); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			next.ServeHTTP(cw, r)
		})
	}
}

func decryptBody(key []byte, r *http.Request) error {
	if r.ContentLength > maxBytes {
		return errContentLengthExceeded
	}

	var content []byte
	var err error
	if r.ContentLength > 0 {
		content = make([]byte, r.ContentLength)
		_, err = io.ReadFull(r.Body, content)
	} else {
		content, err = io.ReadAll(io.LimitReader(r.Body, maxBytes))
	}
	if err != nil {
		return err
	}

	content, err = base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		return err
	}

	output, err := codec.EcbDecrypt(key, content)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.Write(output)
	r.Body = io.NopCloser(&buf)

	return nil
}

type cryptoResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func newCryptoResponseWriter(w http.ResponseWriter) *cryptoResponseWriter {
	return &cryptoResponseWriter{
		ResponseWriter: w,
		buf:            new(bytes.Buffer),
	}
}

func (w *cryptoResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *cryptoResponseWriter) Write(bytes []byte) (int, error) {
	return w.buf.Write(bytes)
}

func (w *cryptoResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

// Hijack 实现 http.Hijacker 接口。
// 如果底层 http.ResponseWriter 支持的话，该方法将扩展响应以满足 http.Hijacker。
func (w *cryptoResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacked, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacked.Hijack()
	}

	return nil, nil, errors.New("服务器不支持 hijacking 劫持")
}

// Flush 发送缓冲数据到客户端。
func (w *cryptoResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *cryptoResponseWriter) flush(key []byte) {
	if w.buf.Len() == 0 {
		return
	}

	content, err := codec.EcbEncrypt(key, w.buf.Bytes())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := base64.StdEncoding.EncodeToString(content)
	if n, err := io.WriteString(w.ResponseWriter, body); err != nil {
		logx.Errorf("写入响应失败，错误：%s", err)
	} else if n < len(content) {
		logx.Errorf("实际字节数：%d，已写入字节数：%d", len(content), n)
	}
}
