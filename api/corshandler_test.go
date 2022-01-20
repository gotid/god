package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorsHandlerWithOrigins(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		expect  string
	}{
		{
			name:   "可以允许所有来源",
			expect: allOrigins,
		},
		{
			name:    "可以允许一个来源",
			origins: []string{"local"},
			expect:  "local",
		},
		{
			name:    "多个来源只有第一个生效",
			origins: []string{"local", "remote"},
			expect:  "local",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler := CorsHandler(test.origins...)
			handler.ServeHTTP(w, nil)
			assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
			assert.Equal(t, test.expect, w.Header().Get(allowOrigin))
		})
	}
}
