package main

import (
	"fmt"

	"github.com/gotid/god/lib/queue/dq"
	"github.com/gotid/god/lib/store/redis"
)

func main() {
	consumer := dq.NewConsumer(dq.Conf{
		Beanstalks: []dq.Beanstalk{
			{
				Endpoint: "dev:11300",
				Tube:     "dhome-sms-login",
			},
			{
				Endpoint: "dev:11301",
				Tube:     "dhome-sms-login",
			},
		},
		Redis: redis.Conf{
			Host: "192.168.0.17:6382",
			Mode: redis.StandaloneMode,
		},
	})

	consumer.Consume(func(body []byte) {
		// time.Sleep(1* time.Second)
		fmt.Println(body)
	})
}
