package syncx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAtomicDuration(t *testing.T) {
	d := ForAtomicDuration(time.Duration(100))
	assert.Equal(t, time.Duration(100), d.Load())
	d.Set(time.Duration(200))
	assert.Equal(t, time.Duration(200), d.Load())
	assert.True(t, d.CompareAndSwap(time.Duration(200), time.Duration(300)))
	assert.Equal(t, time.Duration(300), d.Load())
	assert.False(t, d.CompareAndSwap(time.Duration(200), time.Duration(400)))
	assert.Equal(t, time.Duration(300), d.Load())
}
