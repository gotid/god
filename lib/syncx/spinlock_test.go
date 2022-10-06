package syncx

import (
	"fmt"
	"github.com/gotid/god/lib/lang"
	"github.com/stretchr/testify/assert"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSpinLock_TryLock(t *testing.T) {
	var lock SpinLock
	assert.True(t, lock.TryLock())
	assert.False(t, lock.TryLock())
	lock.Unlock()
	assert.True(t, lock.TryLock())
}

func TestSpinLock(t *testing.T) {
	var lock SpinLock
	lock.Lock()
	assert.False(t, lock.TryLock())
	lock.Unlock()
	assert.True(t, lock.TryLock())
}

func TestSpinLockRace(t *testing.T) {
	var lock SpinLock
	lock.Lock()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
	}()

	time.Sleep(time.Millisecond * 100)
	lock.Unlock()
	wg.Wait()

	assert.True(t, lock.TryLock())
}

func TestSpinLock_TryLock2(t *testing.T) {
	var lock SpinLock
	var count int32
	var wg sync.WaitGroup
	wg.Add(2)
	sig := make(chan lang.PlaceholderType)

	go func() {
		lock.TryLock()
		sig <- lang.Placeholder
		atomic.AddInt32(&count, 1)
		runtime.Gosched()
		lock.Unlock()
		wg.Done()
	}()

	go func() {
		<-sig
		lock.Lock()
		atomic.AddInt32(&count, 1)
		lock.Unlock()
		wg.Done()
	}()

	wg.Wait()
	assert.Equal(t, int32(2), atomic.LoadInt32(&count))
}

func output(s string) {
	for i := 0; i < 3; i++ {
		fmt.Println(s)
	}
}

func Test_GoschedDisabled(t *testing.T) {
	go output("子协程 2")
	output("主协程 1")
}

func Test_GoschedEnabled(t *testing.T) {
	go output("子协程 2")
	runtime.Gosched()
	fmt.Println("我在哪儿")
	output("主协程 1")
}

func Test_GoschedEnableAndSleep(t *testing.T) {
	go func() {
		time.Sleep(10000 * time.Nanosecond)
		output("子协程 2")
	}()
	runtime.Gosched()
	output("主协程 1")
}
