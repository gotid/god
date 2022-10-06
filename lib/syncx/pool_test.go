package syncx

import (
	"github.com/gotid/god/lib/lang"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const limit = 10

func TestPool_Get(t *testing.T) {
	stack := NewPool(limit, create, destroy)
	ch := make(chan lang.PlaceholderType)

	for i := 0; i < limit; i++ {
		var fail AtomicBool
		go func() {
			v := stack.Get()
			if v.(int) != 1 {
				fail.Set(true)
			}
			ch <- lang.Placeholder
		}()

		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Failed()
		}

		if fail.True() {
			t.Fatal("不匹配的值")
		}
	}
}

func TestPoolPopTooMany(t *testing.T) {
	stack := NewPool(limit, create, destroy)
	ch := make(chan lang.PlaceholderType, 1)

	for i := 0; i < limit; i++ {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			stack.Get()
			ch <- lang.Placeholder
			wg.Done()
		}()

		wg.Wait()
		select {
		case <-ch:
		default:
			t.Fail()
		}
	}

	var wg, pushWait sync.WaitGroup
	wg.Add(1)
	pushWait.Add(1)
	go func() {
		pushWait.Done()
		stack.Get()
		wg.Done()
	}()

	pushWait.Wait()
	stack.Put(1)
	wg.Wait()
}

func TestPoolPopFirst(t *testing.T) {
	var value int32
	stack := NewPool(limit, func() interface{} {
		return atomic.AddInt32(&value, 1)
	}, destroy)

	for i := 0; i < 100; i++ {
		v := stack.Get().(int32)
		assert.Equal(t, 1, int(v))
		stack.Put(v)
	}
}

func TestPoolWithMaxAge(t *testing.T) {
	var value int32
	stack := NewPool(limit, func() interface{} {
		return atomic.AddInt32(&value, 1)
	}, destroy, WithMaxAge(time.Millisecond))

	v1 := stack.Get().(int32)
	// put nil should not matter
	stack.Put(nil)
	stack.Put(v1)
	time.Sleep(time.Millisecond * 10)
	v2 := stack.Get().(int32)
	assert.NotEqual(t, v1, v2)
}

func TestNewPoolPanics(t *testing.T) {
	assert.Panics(t, func() {
		NewPool(0, create, destroy)
	})
}

func create() interface{} {
	return 1
}

func destroy(_ interface{}) {

}
