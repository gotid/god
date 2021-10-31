package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.zc0901.com/go/god/lib/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusHandler_Disabled(t *testing.T) {
	promMetricHandler := PrometheusHandler("/user/login")
	handler := promMetricHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestPrometheusHandler_Enabled(t *testing.T) {
	prometheus.StartAgent(prometheus.Config{
		Host: "localhost",
		Path: "/",
	})
	promMetricHandler := PrometheusHandler("/user/login")
	handler := promMetricHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
