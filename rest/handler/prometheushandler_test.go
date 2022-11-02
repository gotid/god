package handler

import (
	"github.com/gotid/god/lib/prometheus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPromMetricHandler_Disabled(t *testing.T) {
	promMetricHandler := PrometheusHandler("/user/login")
	handler := promMetricHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestPromMetricHandler_Enabled(t *testing.T) {
	prometheus.StartAgent(prometheus.Config{
		Host: "localhost",
		Path: "/",
	})
	promMetricHandler := PrometheusHandler("/user/login")
	handler := promMetricHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
