package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/lib/limit"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/store/redis"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// 1秒内，最大速率100，最大突发2
const (
	rate    = 100 // 速率
	burst   = 2   // 突发
	seconds = 1
)

var (
	rds     = flag.String("redis", "localhost:6379", "redis 地址，默认 localhost:6379")
	rdsPass = flag.String("redisPass", "", "redis 密码")
	rdsKey  = flag.String("redisKey", "rate", "redis 键前缀")
	threads = flag.Int("threads", runtime.NumCPU(), "并发线程数，默认为 cpu 个数")
)

func main() {
	flag.Parse()
	logx.Disable()

	store := redis.New(*rds, redis.WithPass(*rdsPass))
	fmt.Println("redis 启动状态：", store.Ping())
	fmt.Println("cpu 内核数：", runtime.NumCPU())

	lmt := limit.NewTokenLimiter(rate, burst, store, *rdsKey)
	timer := time.NewTimer(time.Second * seconds)
	quit := make(chan struct{})
	defer timer.Stop()
	go func() {
		<-timer.C
		close(quit)
	}()

	var allowed, denied int32
	var wait sync.WaitGroup
	for i := 0; i < *threads; i++ {
		wait.Add(1)
		go func() {
			for {
				select {
				case <-quit:
					wait.Done()
					return
				default:
					if lmt.Allow() {
						atomic.AddInt32(&allowed, 1)
					} else {
						atomic.AddInt32(&denied, 1)
					}
				}
			}
		}()
	}
	wait.Wait()

	fmt.Printf("允许：%d，拒绝：%d，每秒请求数：%d\n", allowed, denied, (allowed+denied)/seconds)
}
