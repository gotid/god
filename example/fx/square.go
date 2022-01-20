package main

import (
	"fmt"

	"github.com/gotid/god/lib/fx"
)

func main() {
	result, err := fx.From(func(source chan<- interface{}) {
		for i := 0; i < 5; i++ {
			source <- i
		}
	}).Map(func(item interface{}) interface{} {
		i := item.(int)
		return i * i
	}).Filter(func(item interface{}) bool {
		i := item.(int)
		return i%2 == 0
	}).Reduce(func(pipe <-chan interface{}) (interface{}, error) {
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
