package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTotalWeights(t *testing.T) {
	weights := TotalWeights(Config{
		{Weight: -1},
		{Weight: 0},
		{Weight: 1},
	})

	assert.Equal(t, 1, weights)
}

func TestFormatKeys(t *testing.T) {
	assert.Equal(t, "a,b", formatKeys([]string{"a", "b"}))
}
