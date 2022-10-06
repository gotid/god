package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRefreshCpu(t *testing.T) {
	assert.NotPanics(t, func() {
		RefreshCpu()
	})
}

func BenchmarkRefreshCpu(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RefreshCpu()
	}
}
