package executors

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestBulkExecutor(t *testing.T) {
	var values []int
	var lock sync.Mutex

	executor := NewBulkExecutor(func(tasks []interface{}) {
		lock.Lock()
		values = append(values, len(tasks))
		lock.Unlock()
	}, WithBulkTasks(10), WithBulkInterval(time.Minute))

	for i := 0; i < 50; i++ {
		executor.Add(1)
		time.Sleep(time.Millisecond)
	}

	lock.Lock()
	assert.True(t, len(values) > 0)
	// 忽略最后的值
	for i := 0; i < len(values); i++ {
		assert.Equal(t, 10, values[i])
	}
	lock.Unlock()
}

func TestBulkExecutor_Flush(t *testing.T) {
	const (
		caches = 10
		size   = 5
	)
	var wg sync.WaitGroup

	wg.Add(1)
	executor := NewBulkExecutor(func(tasks []interface{}) {
		assert.Equal(t, size, len(tasks))
		wg.Done()
	}, WithBulkTasks(caches), WithBulkInterval(100*time.Millisecond))

	for i := 0; i < size; i++ {
		executor.Add(1)
	}
	wg.Wait()
}

func TestBulkExecutor_Empty(t *testing.T) {
	NewBulkExecutor(func(tasks []interface{}) {
		assert.Fail(t, "不应该被调用")
	}, WithBulkTasks(10), WithBulkInterval(time.Millisecond))
	time.Sleep(100 * time.Millisecond)
}

func TestBulkExecutorFlush(t *testing.T) {
	const (
		caches = 10
		tasks  = 5
	)

	var wait sync.WaitGroup
	wait.Add(1)
	be := NewBulkExecutor(func(items []interface{}) {
		assert.Equal(t, tasks, len(items))
		wait.Done()
	}, WithBulkTasks(caches), WithBulkInterval(time.Minute))
	for i := 0; i < tasks; i++ {
		be.Add(1)
	}
	be.Flush()
	wait.Wait()
}

func TestBulkExecutorFlushSlowTasks(t *testing.T) {
	const total = 1500
	lock := new(sync.Mutex)
	result := make([]interface{}, 0, 10000)
	be := NewBulkExecutor(func(tasks []interface{}) {
		time.Sleep(time.Millisecond * 100)
		lock.Lock()
		defer lock.Unlock()
		result = append(result, tasks...)
	}, WithBulkTasks(1000))
	for i := 0; i < total; i++ {
		assert.Nil(t, be.Add(i))
	}

	be.Flush()
	be.Wait()
	assert.Equal(t, total, len(result))
}

func BenchmarkBulkExecutor(b *testing.B) {
	b.ReportAllocs()
	be := NewBulkExecutor(func(tasks []interface{}) {
		time.Sleep(time.Millisecond * time.Duration(len(tasks)))
	})
	for i := 0; i < b.N; i++ {
		time.Sleep(200 * time.Microsecond)
		be.Add(1)
	}
	be.Flush()
}
