package stringx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRand(t *testing.T) {
	Seed(time.Now().UnixNano())
	assert.True(t, len(Rand()) > 0)
	assert.True(t, len(RandId()) > 0)
}

func BenchmarkRandN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Randn(10)
	}
}
