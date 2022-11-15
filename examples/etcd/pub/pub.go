package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/lib/discov"
	"log"
	"time"
)

var value = flag.String("v", "value", "发布订阅值")

func main() {
	flag.Parse()

	publisher := discov.NewPublisher([]string{"localhost:2379"}, "028F2C35852D", *value)
	if err := publisher.KeepAlive(); err != nil {
		log.Fatal(err)
	}
	defer publisher.Stop()

	for {
		time.Sleep(time.Second)
		fmt.Println(*value)
	}
}
