package timex

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestRealTicker_DoTick(t *testing.T) {
	ticker := NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	var count int
	for range ticker.Chan() {
		count++
		if count > 5 {
			break
		}
	}
}

func TestFakeTicker(t *testing.T) {
	const total = 5
	ticker := NewFakeTicker()
	defer ticker.Stop()

	var count int32
	go func() {
		for range ticker.Chan() {
			if atomic.AddInt32(&count, 1) == total {
				ticker.Done()
			}
		}
	}()

	for i := 0; i < total; i++ {
		ticker.Tick()
	}

	assert.Nil(t, ticker.Wait(time.Second))
	assert.Equal(t, int32(total), atomic.LoadInt32(&count))
}

func TestFakeTicker_Timeout(t *testing.T) {
	ticker := NewFakeTicker()
	defer ticker.Stop()

	assert.NotNil(t, ticker.Wait(time.Millisecond))
}
