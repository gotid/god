package syncx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLimit(t *testing.T) {
	limit := NewLimit(2)
	limit.Borrow()
	assert.True(t, limit.TryBorrow())
	assert.False(t, limit.TryBorrow())
	assert.Nil(t, limit.Return())
	assert.Nil(t, limit.Return())
	assert.Equal(t, ErrLimitReturn, limit.Return())
}
