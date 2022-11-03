package handler

import (
	"bytes"
	"encoding/base64"
	"github.com/gotid/god/lib/codec"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	reqText  = "ping"
	respText = "pong"
)

var aesKey = []byte(`PdSgVkYp3s6v9y$B&E)H+MbQeThWmZq4`)

func init() {
	log.SetOutput(io.Discard)
}

func TestCryptoHandlerGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/any", http.NoBody)
	handler := CryptoHandler(aesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(respText))
		w.Header().Set("X-Test", "test")
		assert.Nil(t, err)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	expect, err := codec.EcbEncrypt(aesKey, []byte(respText))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "test", recorder.Header().Get("X-Test"))
	assert.Equal(t, base64.StdEncoding.EncodeToString(expect), recorder.Body.String())
}

func TestCryptoHandlerPost(t *testing.T) {
	var buf bytes.Buffer
	enc, err := codec.EcbEncrypt(aesKey, []byte(reqText))
	assert.Nil(t, err)
	buf.WriteString(base64.StdEncoding.EncodeToString(enc))

	req := httptest.NewRequest(http.MethodPost, "/any", &buf)
	handler := CryptoHandler(aesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		assert.Equal(t, reqText, string(body))

		w.Write([]byte(respText))
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	expect, err := codec.EcbEncrypt(aesKey, []byte(respText))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, base64.StdEncoding.EncodeToString(expect), recorder.Body.String())
}

func TestCryptoHandlerPostBadEncryption(t *testing.T) {
	var buf bytes.Buffer
	enc, err := codec.EcbEncrypt(aesKey, []byte(reqText))
	assert.Nil(t, err)
	buf.Write(enc)

	req := httptest.NewRequest(http.MethodPost, "/any", &buf)
	handler := CryptoHandler(aesKey)(nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCryptoHandlerWriteHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/any", http.NoBody)
	handler := CryptoHandler(aesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestCryptoHandlerFlush(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/any", http.NoBody)
	handler := CryptoHandler(aesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(respText))
		flusher, ok := w.(http.Flusher)
		assert.True(t, ok)
		flusher.Flush()
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	expect, err := codec.EcbEncrypt(aesKey, []byte(respText))
	assert.Nil(t, err)
	assert.Equal(t, base64.StdEncoding.EncodeToString(expect), recorder.Body.String())
}

func TestCryptoHandler_Hijack(t *testing.T) {
	resp := httptest.NewRecorder()
	writer := newCryptoResponseWriter(resp)
	assert.NotPanics(t, func() {
		writer.Hijack()
	})

	writer = newCryptoResponseWriter(mockedHijackable{resp})
	assert.NotPanics(t, func() {
		writer.Hijack()
	})
}

func TestCryptoHandler_ContentTooLong(t *testing.T) {
	handler := CryptoHandler(aesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	svr := httptest.NewServer(handler)
	defer svr.Close()

	body := make([]byte, maxBytes+1)
	rand.Read(body)
	req, err := http.NewRequest(http.MethodPost, svr.URL, bytes.NewReader(body))
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
