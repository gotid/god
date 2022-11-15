package fx

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestParallel(t *testing.T) {
	var count int32
	Parallel(func() {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&count, 1)
	}, func() {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&count, 1)
	}, func() {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&count, 1)
	})
	assert.Equal(t, int32(3), count)
}
