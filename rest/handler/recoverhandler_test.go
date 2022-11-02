package handler

import (
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	log.SetOutput(io.Discard)
}

func TestWithPanic(t *testing.T) {
	handler := RecoverHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("whatever")
	}))

	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestWithoutPanic(t *testing.T) {
	handler := RecoverHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))

	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
