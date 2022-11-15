package main

import (
	"fmt"
	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/logx"
	"time"
)

func main() {
	sub, err := discov.NewSubscriber([]string{"localhost:2379"}, "028F2C35852D", discov.Exclusive())
	logx.Must(err)

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("值：", sub.Values())
		}
	}
}
