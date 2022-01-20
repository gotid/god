package handler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gotid/god/lib/codec"

	"github.com/gotid/god/api/httpx"
	"github.com/stretchr/testify/assert"
)

func TestGunzipHandler(t *testing.T) {
	t.Run("压缩测试", func(t *testing.T) {
		const message = "hello world"
		var wg sync.WaitGroup
		wg.Add(1)
		handler := GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.Nil(t, err)
			assert.Equal(t, string(body), message)
			wg.Done()
		}))

		req := httptest.NewRequest(http.MethodPost, "http://localhost",
			bytes.NewReader(codec.Gzip([]byte(message))))
		req.Header.Set(httpx.ContentEncoding, gzipEncoding)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		wg.Wait()
	})

	t.Run("无压缩", func(t *testing.T) {
		const message = "hello world"
		var wg sync.WaitGroup
		wg.Add(1)
		handler := GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.Nil(t, err)
			assert.Equal(t, string(body), message)
			wg.Done()
		}))

		req := httptest.NewRequest(http.MethodPost, "http://localhost",
			strings.NewReader(message))
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		wg.Wait()
	})

	t.Run("无压缩但会告知错误", func(t *testing.T) {
		const message = "hello world"
		handler := GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		req := httptest.NewRequest(http.MethodPost, "http://localhost",
			strings.NewReader(message))
		req.Header.Set(httpx.ContentEncoding, gzipEncoding)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}
