package iox

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBufferPool(t *testing.T) {
	capacity := 10
	pool := NewBufferPool(capacity)
	pool.Put(bytes.NewBuffer(make([]byte, 0, 2*capacity)))
	assert.True(t, pool.Get().Cap() <= capacity)
}
