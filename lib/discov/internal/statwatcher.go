//go:generate mockgen -package internal -destination statewatcher_mock.go -source statewatcher.go etcdConn

package internal

import (
	"context"
	"google.golang.org/grpc/connectivity"
	"sync"
)

type (
	etcdConn interface {
		GetState() connectivity.State
		WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool
	}

	stateWatcher struct {
		disconnected bool
		currentState connectivity.State
		listeners    []func()
		// 用于保护监听器列表，因为只有 listeners 能被其他协程访问。
		lock sync.Mutex
	}
)

func newStateWatcher() *stateWatcher {
	return new(stateWatcher)
}

func (w *stateWatcher) addListener(l func()) {
	w.lock.Lock()
	w.listeners = append(w.listeners, l)
	w.lock.Unlock()
}

func (w *stateWatcher) watch(conn etcdConn) {
	w.currentState = conn.GetState()
	for {
		if conn.WaitForStateChange(context.Background(), w.currentState) {
			w.updateState(conn)
		}
	}
}

func (w *stateWatcher) updateState(conn etcdConn) {
	w.currentState = conn.GetState()
	switch w.currentState {
	case connectivity.TransientFailure, connectivity.Shutdown:
		w.disconnected = true
	case connectivity.Ready:
		if w.disconnected {
			w.disconnected = false
			w.notifyListeners()
		}
	}
}

func (w *stateWatcher) notifyListeners() {
	w.lock.Lock()
	defer w.lock.Unlock()

	for _, l := range w.listeners {
		l()
	}
}
