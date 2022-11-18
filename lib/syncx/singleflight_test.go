package syncx

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExclusiveCallDo(t *testing.T) {
	f := NewSingleFlight()
	v, err := f.Do("key", func() (any, error) {
		return "bar", nil
	})
	if got, want := fmt.Sprintf("%v (%T)", v, v), "bar (string)"; got != want {
		t.Errorf("Do = %v, want %v", got, want)
	}
	if err != nil {
		t.Errorf("Do error = %v", err)
	}
}

func TestExclusiveCallDoErr(t *testing.T) {
	f := NewSingleFlight()
	someErr := errors.New("some error")
	v, err := f.Do("key", func() (any, error) {
		return nil, someErr
	})
	if err != someErr {
		t.Errorf("Do error = %v, want someErr", err)
	}
	if v != nil {
		t.Errorf("不期待的非空值 %#v", v)
	}
}

func TestExclusiveCallDoDupSuppress(t *testing.T) {
	f := NewSingleFlight()
	c := make(chan string)
	var calls int32
	fn := func() (any, error) {
		atomic.AddInt32(&calls, 1)
		return <-c, nil
	}

	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			v, err := f.Do("key", fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
			}
			if v.(string) != "bar" {
				t.Errorf("got %q; want %q", v, "bar")
			}
			wg.Done()
		}()
	}
	time.Sleep(100 * time.Millisecond) // 阻塞上述协程
	c <- "bar"
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("实际调用次数 = %d; 期望 1", got)
	}
}

func TestExclusiveCallDoDiffDupSuppress(t *testing.T) {
	f := NewSingleFlight()
	broadcast := make(chan struct{})
	var calls int32
	tests := []string{"c", "a", "c", "a", "b"}

	var wg sync.WaitGroup
	for _, key := range tests {
		wg.Add(1)
		go func(k string) {
			<-broadcast // 准备好所有协程
			_, err := f.Do(k, func() (any, error) {
				atomic.AddInt32(&calls, 1)
				time.Sleep(10 * time.Millisecond)
				return nil, nil
			})
			if err != nil {
				t.Errorf("Do error: %v", err)
			}
			wg.Done()
		}(key)
	}
	time.Sleep(100 * time.Millisecond) // 阻塞上述协程
	close(broadcast)
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("实际调用次数 = %d; 期望 3", got)
	}
}

func TestExclusiveCallDoExDupSuppress(t *testing.T) {
	f := NewSingleFlight()
	c := make(chan string)
	var calls int32
	fn := func() (any, error) {
		atomic.AddInt32(&calls, 1)
		return <-c, nil
	}

	const n = 10
	var wg sync.WaitGroup
	var freshes int32
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			v, fresh, err := f.DoEx("key", fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
			}
			if fresh {
				atomic.AddInt32(&freshes, 1)
			}
			if v.(string) != "bar" {
				t.Errorf("got %q; want %q", v, "bar")
			}
			wg.Done()
		}()
	}
	time.Sleep(100 * time.Millisecond) // 阻塞上述协程
	c <- "bar"
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("实际调用次数 = %d，期望 1", got)
	}
	if got := atomic.LoadInt32(&freshes); got != 1 {
		t.Errorf("实际新鲜调用 = %d，期望 1", got)
	}
}
