package syncx

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLockedCalls_Do(t *testing.T) {
	calls := NewLockedCalls()
	v, err := calls.Do("key", func() (any, error) {
		return "bar", nil
	})
	if got, want := fmt.Sprintf("%v (%T)", v, v), "bar (string)"; got != want {
		t.Errorf("Do = %v; want %v", got, want)
	}
	if err != nil {
		t.Errorf("Do error = %v", err)
	}
}

func TestLockedCalls_Err(t *testing.T) {
	calls := NewLockedCalls()
	someErr := errors.New("some error")
	v, err := calls.Do("key", func() (any, error) {
		return nil, someErr
	})
	if err != someErr {
		t.Errorf("Do error = %v; want someError", err)
	}
	if v != nil {
		t.Errorf("不期待的非空值 %#v", v)
	}
}

func TestNewLockedCalls(t *testing.T) {
	lc := NewLockedCalls()
	c := make(chan string)
	var calls int
	fn := func() (any, error) {
		calls++
		ret := calls
		<-c
		calls--
		return ret, nil
	}

	const n = 10
	var results []int
	var lock sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			v, err := lc.Do("key", fn)
			if err != nil {
				t.Errorf("Do err: %v", err)
			}

			lock.Lock()
			results = append(results, v.(int))
			lock.Unlock()
			wg.Done()
		}()
	}

	time.Sleep(100 * time.Millisecond) // 阻塞上方子协程
	for i := 0; i < n; i++ {
		c <- "bar"
	}
	wg.Wait()

	lock.Lock()
	defer lock.Unlock()

	for _, item := range results {
		if item != 1 {
			t.Errorf("number of calls = %d; want 1", item)
		}
	}
}
