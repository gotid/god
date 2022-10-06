package syncx

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestTimeoutLimit(t *testing.T) {
	limit := NewTimeoutLimit(2)
	assert.Nil(t, limit.Borrow(time.Millisecond*200))
	assert.Nil(t, limit.Borrow(time.Millisecond*200))

	var wg1, wg2, wg3 sync.WaitGroup
	wg1.Add(1)
	wg2.Add(1)
	wg3.Add(1)
	go func() {
		wg1.Wait()
		wg2.Done()
		assert.Nil(t, limit.Return())
		wg3.Done()
	}()
	wg1.Done()
	wg2.Wait()
	assert.Nil(t, limit.Borrow(time.Second))
	wg3.Wait()
	assert.Equal(t, ErrTimeout, limit.Borrow(time.Millisecond*100))
	assert.Nil(t, limit.Return())
	assert.Nil(t, limit.Return())
	assert.Equal(t, ErrLimitReturn, limit.Return())
}
