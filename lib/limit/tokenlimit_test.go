package limit

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/lib/store/redis/redistest"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func init() {
	logx.Disable()
}

func TestTokenLimit_WithCtx(t *testing.T) {
	s, err := miniredis.Run()
	assert.Nil(t, err)

	const (
		rate  = 5
		burst = 10

		total = 100
	)
	l := NewTokenLimiter(rate, burst, redis.New(s.Addr()), "tokenLimit")
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())
	ok := l.AllowCtx(ctx)
	assert.True(t, ok)

	cancel()
	for i := 0; i < total; i++ {
		ok := l.AllowCtx(ctx)
		assert.False(t, ok)
		assert.False(t, l.monitorStarted)
	}
}

func TestTokenLimit_Rescue(t *testing.T) {
	s, err := miniredis.Run()
	assert.Nil(t, err)

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := NewTokenLimiter(rate, burst, redis.New(s.Addr()), "tokenLimit")
	s.Close()

	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Second / time.Duration(total))
		if i == total>>1 {
			assert.Nil(t, s.Restart())
		}
		if l.Allow() {
			allowed++
		}

		// make sure start monitor more than once doesn't matter
		l.startMonitor()
	}

	assert.True(t, allowed >= burst+rate)
}

func TestTokenLimit_Take(t *testing.T) {
	store, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := NewTokenLimiter(rate, burst, store, "tokenLimit")
	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Second / time.Duration(total))
		if l.Allow() {
			allowed++
		}
	}

	assert.True(t, allowed >= burst+rate)
}

func TestTokenLimit_TakeBurst(t *testing.T) {
	store, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := NewTokenLimiter(rate, burst, store, "tokenLimit")
	var allowed int
	for i := 0; i < total; i++ {
		if l.Allow() {
			allowed++
		}
	}

	assert.True(t, allowed >= burst)
}
