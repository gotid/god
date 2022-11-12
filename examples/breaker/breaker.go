package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/breaker"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/logx"
	"gopkg.in/cheggaaa/pb.v1"
)

const (
	breakRange = 20
	workRange  = 50
	duration   = 5 * time.Minute

	requestInterval = time.Millisecond
	stateFator      = float64(time.Second/requestInterval) / 2
)

type (
	metric struct {
		calls int64
	}

	server struct {
		state int32
	}
)

func (m *metric) addCall() {
	atomic.AddInt64(&m.calls, 1)
}

func (m *metric) reset() int64 {
	return atomic.SwapInt64(&m.calls, 0)
}

func (s *server) serve(m *metric) bool {
	m.addCall()
	return atomic.LoadInt32(&s.state) == 1
}

func (s *server) start() {
	go func() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		var state int32
		for {
			var v int32
			if state == 0 {
				v = r.Int31n(breakRange)
			} else {
				v = r.Int31n(workRange)
			}
			time.Sleep(time.Second * time.Duration(v+1))
			state ^= 1
			atomic.StoreInt32(&s.state, state)
		}
	}()
}

func newServer() *server {
	return &server{}
}

func runBreaker(srv *server, brk breaker.Breaker, duration time.Duration, m *metric) {
	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()
	done := make(chan lang.PlaceholderType)

	go func() {
		time.Sleep(duration)
		close(done)
	}()

	for {
		select {
		case <-ticker.C:
			_ = brk.Do(func() error {
				if srv.serve(m) {
					return nil
				} else {
					return breaker.ErrServiceUnavailable
				}
			})
		case <-done:
			return
		}
	}
}

func main() {
	srv := newServer()
	srv.start()

	brk := breaker.New()
	fp, err := os.Create("result.csv")
	logx.Must(err)
	defer fp.Close()
	fmt.Fprintln(fp, "seconds,state,googleCalls,netflixCalls")

	var gm, nm metric
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		var seconds int
		for range ticker.C {
			seconds++
			gcalls := gm.reset()
			ncalls := nm.reset()
			fmt.Fprintf(fp, "%d,%.2f,%d,%d\n",
				seconds, float64(atomic.LoadInt32(&srv.state))*stateFator, gcalls, ncalls)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		runBreaker(srv, brk, duration, &gm)
		wg.Done()
	}()

	go func() {
		bar := pb.New(int(duration / time.Second)).Start()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			bar.Increment()
		}
		bar.Finish()
	}()

	wg.Wait()
}
