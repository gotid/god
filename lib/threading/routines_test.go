package threading

import (
	"fmt"
	"github.com/gotid/god/lib/lang"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunSafe(t *testing.T) {
	//log.SetOutput(io.Discard)

	i := 0
	defer func() {
		assert.Equal(t, 1, i)
	}()

	ch := make(chan lang.PlaceholderType)
	go RunSafe(func() {
		defer func() {
			ch <- lang.Placeholder
		}()

		panic("panic啦...")
	})

	<-ch
	i++
}

func TestRoutineId(t *testing.T) {
	id := RoutineId()
	fmt.Println(id)
	assert.True(t, id > 0)
}
