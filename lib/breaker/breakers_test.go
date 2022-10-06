package breaker

import (
	"fmt"
	"github.com/gotid/god/lib/stat"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	stat.SetReporter(nil)
}

func TestBreakers(t *testing.T) {
	assert.Nil(t, Do("any", func() error {
		return nil
	}))
}

func verify(t *testing.T, fn func() bool) {
	var count int
	for i := 0; i < 100; i++ {
		if fn() {
			count++
		}
	}
	assert.True(t, count >= 80, fmt.Sprintf("应大于80，实际为 %d", count))
}
