package health

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const probeName = "probe"

func TestHealthManager(t *testing.T) {
	hm := NewHealthManager(probeName)
	assert.False(t, hm.IsReady())

	hm.MarkReady()
	assert.True(t, hm.IsReady())
}