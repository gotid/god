package handler

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/gotid/god/lib/codec"
	"github.com/gotid/god/lib/logx"
)

const maxBytes = 1 << 20 // 1MB

var errContentLengthExceeded = errors.New("密文长度超出限制")

// CryptionHandler 返回一个处理加密的中间件。
func CryptionHandler(key []byte) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ew := newCryptionResponseWriter(w)
			defer ew.flush(key)

			if r.ContentLength <= 0 {
				next.ServeHTTP(ew, r)
				return
			}

			if err := decryptBody(key, r); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			next.ServeHTTP(ew, r)
		})
	}
}

// 加密响应输出器
type cryptionResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (w *cryptionResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *cryptionResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacked, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacked.Hijack()
	}

	return nil, nil, errors.New("服务器不支持 Hijacker")
}

func (w *cryptionResponseWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *cryptionResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *cryptionResponseWriter) flush(key []byte) {
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
		logx.Errorf("写响应失败，错误：%s", err)
	} else if n < len(content) {
		logx.Errorf("实际字节数: %d，写字节数: %d", len(content), n)
	}
}

func newCryptionResponseWriter(w http.ResponseWriter) *cryptionResponseWriter {
	return &cryptionResponseWriter{
		ResponseWriter: w,
		buf:            new(bytes.Buffer),
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
		content, err = ioutil.ReadAll(io.LimitReader(r.Body, maxBytes))
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
	r.Body = ioutil.NopCloser(&buf)

	return nil
}
