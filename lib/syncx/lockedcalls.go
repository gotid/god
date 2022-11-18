package syncx

import "sync"

type (
	// LockedCalls 确保相同key的调用按照顺序执行。
	// 例如，A 调用 F，在完成前 B 也调用 F，那么 B 的调用不会被阻止，
	// 在 A 调用完成后，B 的调用再开始执行。
	LockedCalls interface {
		Do(key string, fn func() (any, error)) (any, error)
	}

	lockedGroup struct {
		mu sync.Mutex
		m  map[string]*sync.WaitGroup
	}
)

// NewLockedCalls 返回一个 LockedCalls。
func NewLockedCalls() LockedCalls {
	return &lockedGroup{
		m: make(map[string]*sync.WaitGroup),
	}
}

func (lg *lockedGroup) Do(key string, fn func() (any, error)) (any, error) {
begin:
	lg.mu.Lock()
	if wg, ok := lg.m[key]; ok {
		lg.mu.Unlock()
		wg.Wait()
		goto begin
	}

	return lg.makeCall(key, fn)
}

func (lg *lockedGroup) makeCall(key string, fn func() (any, error)) (any, error) {
	var wg sync.WaitGroup
	wg.Add(1)
	lg.m[key] = &wg
	lg.mu.Unlock()

	defer func() {
		// 先删键，再Done。
		// 顺序不能反，否则另一个 Do 调用会一直 Wait 而收不到 Done 的通知。
		lg.mu.Lock()
		delete(lg.m, key)
		lg.mu.Unlock()
		wg.Done()
	}()

	return fn()
}
