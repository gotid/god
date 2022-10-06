package threading

import (
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRoutineGroup_Run(t *testing.T) {
	var count int32
	group := NewRoutineGroup()
	for i := 0; i < 3; i++ {
		group.Run(func() {
			atomic.AddInt32(&count, 1)
		})
	}
	group.Wait()

	assert.Equal(t, int32(3), count)
}

func TestRoutineGroup_RunSafe(t *testing.T) {
	log.SetOutput(io.Discard)

	var count int32
	group := NewRoutineGroup()
	var once sync.Once
	for i := 0; i < 3; i++ {
		group.RunSafe(func() {
			once.Do(func() {
				panic("")
			})
			atomic.AddInt32(&count, 1)
		})
	}

	group.Wait()

	assert.Equal(t, int32(2), count)
}
