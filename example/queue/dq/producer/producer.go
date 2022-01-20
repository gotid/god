package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gotid/god/lib/queue/dq"
)

func main() {
	producer := dq.NewProducer([]dq.Beanstalk{
		{
			Endpoint: "dev:11300",
			Tube:     "dhome-sms-login",
		},
		{
			Endpoint: "dev:11301",
			Tube:     "dhome-sms-login",
		},
	})

	for i := 0; i < 1; i++ {
		// time.Sleep(time.Duration(1))
		id, err := producer.Delay([]byte("18301365447,1234"), time.Second*0)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		fmt.Println("job id", id)
	}
}
