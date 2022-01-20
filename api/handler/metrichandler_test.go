package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.zc0901.com/go/god/lib/stat"
	"github.com/stretchr/testify/assert"
)

func TestMetricHandler(t *testing.T) {
	metrics := stat.NewMetrics("单元测试")
	metricHandler := MetricHandler(metrics)
	handler := metricHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
