package syncx

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestNewAtomicFloat64(t *testing.T) {
	f := ForAtomicFloat64(100)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				f.Add(1)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	assert.Equal(t, float64(600), f.Load())
}
