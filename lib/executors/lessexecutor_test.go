package executors

import (
	"github.com/gotid/god/lib/timex"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLessExecutor_DoOrDiscard(t *testing.T) {
	executor := NewLessExecutor(time.Minute)
	assert.True(t, executor.DoOrDiscard(func() {}))
	assert.False(t, executor.DoOrDiscard(func() {}))
	executor.lastTime.Set(timex.Now() - time.Minute - 30*time.Second)
	assert.True(t, executor.DoOrDiscard(func() {}))
	assert.False(t, executor.DoOrDiscard(func() {}))
}

func BenchmarkLessExecutor_DoOrDiscard(b *testing.B) {
	b.ReportAllocs()
	executor := NewLessExecutor(time.Millisecond)
	for i := 0; i < b.N; i++ {
		executor.DoOrDiscard(func() {})
	}
}
