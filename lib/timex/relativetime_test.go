package timex

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRelativeTime(t *testing.T) {
	time.Sleep(time.Millisecond)
	now := Now()
	assert.True(t, now > 0)
	time.Sleep(time.Millisecond)
	assert.True(t, Since(now) > 0)
}

func BenchmarkTimeSince(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = time.Since(time.Now())
	}
}

func BenchmarkTimeXSince(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = Since(Now())
	}
}
