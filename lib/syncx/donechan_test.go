package syncx

import (
	"sync"
	"testing"
)

func TestDoneChan_Close(t *testing.T) {
	doneChan := NewDoneChan()
	for i := 0; i < 5; i++ {
		doneChan.Close()
	}
}

func TestDoneChan_Done(t *testing.T) {
	var wg sync.WaitGroup
	doneChan := NewDoneChan()

	wg.Add(1)
	go func() {
		<-doneChan.Done()
		wg.Done()
	}()

	for i := 0; i < 5; i++ {
		doneChan.Close()
	}

	wg.Wait()
}
