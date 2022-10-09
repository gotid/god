package threading

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestTaskRunner_Schedule(t *testing.T) {
	times := 100
	runner := NewTaskRunner(runtime.NumCPU())

	var counter int32
	var wg sync.WaitGroup
	for i := 0; i < times; i++ {
		wg.Add(1)
		runner.Schedule(func() {
			atomic.AddInt32(&counter, 1)
			wg.Done()
		})
	}
	wg.Wait()

	assert.Equal(t, times, int(counter))
}
