package main

import (
	"fmt"
	"github.com/gotid/god/lib/stringx"
	"github.com/gotid/god/lib/syncx"
	"sync"
	"time"
)

// 并发5个协程同时获取随机编号
func main() {
	const round = 5
	var wg sync.WaitGroup
	flight := syncx.NewSingleFlight()

	wg.Add(round)
	for i := 0; i < round; i++ {
		go func() {
			defer wg.Done()
			val, err := flight.Do("once", func() (any, error) {
				time.Sleep(time.Second)
				fmt.Println("该函数在并发时，只会被调用和打印一次")
				return stringx.RandId(), nil
			})
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(val)
			}
		}()
	}
	wg.Wait()
}
