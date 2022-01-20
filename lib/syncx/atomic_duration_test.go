package syncx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAtomicDuration_Load(t *testing.T) {
	ad := ForAtomicDuration(time.Duration(100))
	assert.Equal(t, time.Duration(100), ad.Load())
}

func TestAtomicDuration_Set(t *testing.T) {
	ad := NewAtomicDuration()
	ad.Set(time.Duration(200))
	assert.Equal(t, time.Duration(200), ad.Load())
}

func TestAtomicDuration_CompareAndSwap(t *testing.T) {
	ad := ForAtomicDuration(time.Duration(200))

	assert.True(t, ad.CompareAndSwap(time.Duration(200), time.Duration(300)))
	assert.False(t, ad.CompareAndSwap(time.Duration(200), time.Duration(400)))
	assert.Equal(t, time.Duration(300), ad.Load())
}
