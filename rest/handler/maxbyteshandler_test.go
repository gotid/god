package handler

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMaxBytesHandler(t *testing.T) {
	mbh := MaxBytesHandler(10)
	handler := mbh(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodPost, "http://localhost",
		bytes.NewBufferString("123456789012345"))
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.Code)

	req = httptest.NewRequest(http.MethodPost, "http://localhost", bytes.NewBufferString("12345"))
	resp = httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestMaxBytesHandlerNoLimit(t *testing.T) {
	mbh := MaxBytesHandler(-1)
	handler := mbh(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodPost, "http://localhost",
		bytes.NewBufferString("123456789012345"))
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
