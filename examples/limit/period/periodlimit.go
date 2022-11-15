package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/lib/limit"
	"github.com/gotid/god/lib/store/redis"
	"log"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const seconds = 5

var (
	rds     = flag.String("redis", "localhost:6379", "redis 地址，默认 localhost:6379")
	rdsPass = flag.String("redisPass", "", "redis 密码")
	rdsKey  = flag.String("redisKey", "rate", "redis 键前缀")
	threads = flag.Int("threads", runtime.NumCPU(), "并发线程数，默认为 cpu 个数")
)

func main() {
	flag.Parse()

	store := redis.New(*rds, redis.WithPass(*rdsPass))
	fmt.Println("redis 启动状态：", store.Ping())
	fmt.Println("cpu 内核数：", runtime.NumCPU())

	lmt := limit.NewPeriodLimit(seconds, 5, store, *rdsKey)
	timer := time.NewTimer(seconds * time.Second)
	quit := make(chan struct{})
	defer timer.Stop()
	go func() {
		<-timer.C
		close(quit)
	}()

	var allowed, denied int32
	var wg sync.WaitGroup
	for i := 0; i < *threads; i++ {
		i := i
		wg.Add(1)
		go func() {
			for {
				select {
				case <-quit:
					wg.Done()
					return
				default:
					if v, err := lmt.Take(strconv.FormatInt(int64(i), 10)); err == nil && v == limit.Allowed {
						atomic.AddInt32(&allowed, 1)
					} else if err != nil {
						log.Fatal(err)
					} else {
						atomic.AddInt32(&denied, 1)
					}
				}
			}
		}()
	}
	wg.Wait()

	fmt.Printf("允许：%d，拒绝：%d，每秒请求数：%d\n", allowed, denied, (allowed+denied)/seconds)
}
