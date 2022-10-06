package syncx

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestCond_Wait(t *testing.T) {
	var wg sync.WaitGroup
	cond := NewCond()
	wg.Add(2)
	go func() {
		cond.Wait()
		wg.Done()
	}()
	time.Sleep(time.Duration(50) * time.Millisecond)
	go func() {
		cond.Signal()
		wg.Done()
	}()
	wg.Wait()
}

func TestCond_WaitWithTimeout(t *testing.T) {
	var wg sync.WaitGroup
	cond := NewCond()
	wg.Add(1)
	go func() {
		cond.WaitWithTimeout(time.Duration(500) * time.Millisecond)
		wg.Done()
	}()
	wg.Wait()
}

func TestCond_WaitWithTimeout2(t *testing.T) {
	var wg sync.WaitGroup
	cond := NewCond()
	wg.Add(2)
	ch := make(chan time.Duration, 1)
	defer close(ch)

	timeout := time.Duration(2000) * time.Millisecond
	go func() {
		remainTimeout, _ := cond.WaitWithTimeout(timeout)
		ch <- remainTimeout
		wg.Done()
	}()
	time.Sleep(time.Duration(200) * time.Millisecond)
	go func() {
		cond.Signal()
		wg.Done()
	}()
	wg.Wait()

	remainTimeout := <-ch
	assert.True(t, remainTimeout < timeout, "期望 remainTimeout %v < %v", remainTimeout, timeout)
	assert.True(t, remainTimeout >= time.Duration(200)*time.Millisecond, "期望 remainTimeout %v >= 200毫秒", remainTimeout)
}

func TestSignalNoWait(t *testing.T) {
	cond := NewCond()
	cond.Signal()
}
