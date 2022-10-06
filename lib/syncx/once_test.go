package syncx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOnce(t *testing.T) {
	var v int
	add := Once(func() {
		v++
	})

	for i := 0; i < 5; i++ {
		add()
	}

	assert.Equal(t, 1, v)
}

func BenchmarkOnce(b *testing.B) {
	var v int
	add := Once(func() {
		v++
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		add()
	}
	assert.Equal(b, 1, v)
}
