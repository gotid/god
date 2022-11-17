package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/lib/executors"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"gopkg.in/cheggaaa/pb.v1"
)

const (
	interval = time.Second
	total    = 400
	factor   = 5
	beta     = 0.9
)

var (
	seconds             = flag.Int("d", 10, "持续时间")
	lessWriter          *executors.LessExecutor
	index               int32
	flying              uint64
	aggressiveLock      syncx.SpinLock
	avgFlyingAggressive float64

	bothLock      syncx.SpinLock
	avgFlyingBoth float64

	passCounter = collection.NewRollingWindow(50, time.Millisecond*100)
	rtCounter   = collection.NewRollingWindow(50, time.Millisecond*100)

	lazyLock      syncx.SpinLock
	avgFlyingLazy float64
)

func main() {
	flag.Parse()

	// 只记录 100 条日志
	lessWriter = executors.NewLessExecutor(interval * total / 100)

	fp, err := os.Create("result.csv")
	logx.Must(err)
	defer fp.Close()
	fmt.Fprintln(fp, "second,maxFlight,flying,aggressiveAvgFlying,lazyAvgFlying,bothAvgFlying")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	bar := pb.New(*seconds * 2).Start()
	var wg sync.WaitGroup
	batchRequests := func(i int) {
		<-ticker.C
		requests := (i + 1) * factor
		func() {
			it := time.NewTicker(interval / time.Duration(requests))
			defer it.Stop()
			for j := 0; j < requests; j++ {
				<-it.C
				wg.Add(1)
				go func() {
					issueRequest(fp, atomic.AddInt32(&index, 1))
					wg.Done()
				}()
			}
			bar.Increment()
		}()
	}

	// 请求数越来越多
	for i := 0; i < *seconds; i++ {
		batchRequests(i)
	}

	// 请求数越来越少
	for i := *seconds; i > 0; i-- {
		batchRequests(i)
	}
	bar.Finish()
	wg.Wait()

}

func issueRequest(writer io.Writer, idx int32) {
	v := atomic.AddUint64(&flying, 1)
	aggressiveLock.Lock()
	af := avgFlyingAggressive*beta + float64(v)*(1-beta) // 在飞均值
	avgFlyingAggressive = af
	aggressiveLock.Unlock()

	bothLock.Lock()
	bf := avgFlyingBoth*beta + float64(v)*(1-beta)
	avgFlyingBoth = bf
	bothLock.Unlock()

	duration := time.Millisecond * time.Duration(rand.Int63n(2)+1)
	job(duration) // 模拟一项耗时执行的工作
	passCounter.Add(1)
	rtCounter.Add(float64(duration) / float64(time.Millisecond))

	v1 := atomic.AddUint64(&flying, ^uint64(0))
	lazyLock.Lock()
	lf := avgFlyingLazy*beta + float64(v1)*(1-beta)
	avgFlyingLazy = lf
	lazyLock.Unlock()

	bothLock.Lock()
	bf = avgFlyingBoth*beta + float64(v1)*(1-beta)
	avgFlyingBoth = bf
	bothLock.Unlock()

	lessWriter.DoOrDiscard(func() {
		fmt.Fprintf(writer, "%d,%d,%d,%.2f,%.2f,%.2f\n", idx, maxFlight(), v, af, lf, bf)
	})
}

func maxFlight() int64 {
	return int64(math.Max(1, float64(maxPass()*10)*(minRt()/1e3)))
}

func minRt() float64 {
	var result float64 = 1000

	rtCounter.Reduce(func(b *collection.Bucket) {
		if b.Count <= 0 {
			return
		}

		avg := math.Round(b.Sum / float64(b.Count))
		if avg < result {
			result = avg
		}
	})

	return result
}

func maxPass() int64 {
	var result float64 = 1

	passCounter.Reduce(func(b *collection.Bucket) {
		if b.Sum > result {
			result = b.Sum
		}
	})

	return int64(result)
}

func job(duration time.Duration) {
	time.Sleep(duration)
}
