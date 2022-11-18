package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/lib/collection"
	"log"
	"runtime"
	"sync/atomic"
	"time"
)

const interval = time.Second * 10

var traditional = flag.Bool("traditional", false, "是否启用传统模式，默认否。")

func main() {
	flag.Parse()

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf("当前协程数：%d\n", runtime.NumGoroutine())
			}
		}
	}()

	if *traditional {
		traditionalMode()
	} else {
		timingWheelMode()
	}
}

// 传统模式，协程数量会越来越多
func traditionalMode() {
	var count uint64
	for {
		go func() {
			timer := time.NewTimer(interval)
			defer timer.Stop()

			select {
			case <-timer.C:
				job(&count, nil, nil)
			}
		}()

		time.Sleep(time.Millisecond)
	}
}

// 时间轮模式，协程数量持续稳定在低位
func timingWheelMode() {
	var count uint64
	tw, err := collection.NewTimingWheel(time.Second, 600, func(key, val any) {
		job(&count, key, val)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tw.Stop()

	for i := 0; ; i++ {
		err = tw.SetTimer(i, i, interval)
		if err != nil {
			return
		}
		time.Sleep(time.Microsecond)
	}
}

func job(count *uint64, key any, val any) {
	v := atomic.AddUint64(count, 1)
	if v%10000 == 0 {
		fmt.Println(v, key, val)
	}
}
