package main

import (
	"fmt"
	"testing"

	"github.com/gotid/god/lib/fx"
)

func TestFxSplit(t *testing.T) {
	fx.Just(1, 2, 3, 4, 5).Split(2).ForEach(func(item interface{}) {
		vals := item.([]interface{})
		fmt.Println(vals)
	})
}
