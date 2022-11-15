package main

import (
	"fmt"
	"github.com/gotid/god/lib/fx"
	"time"
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Println("耗时 ", time.Since(start))
	}()

	result, err := fx.From(func(source chan<- any) {
		for i := 0; i < 10; i++ {
			source <- i
		}
	}).Map(func(item any) any {
		i := item.(int)
		return i * i
	}).Filter(func(item any) bool {
		i := item.(int)
		return i%2 == 0
	}).Distinct(func(item any) any {
		return item
	}).Reduce(func(pipe <-chan any) (any, error) {
		var result int
		for item := range pipe {
			i := item.(int)
			result += i
		}
		return result, nil
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}

}
