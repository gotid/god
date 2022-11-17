package main

import (
	"fmt"
	"github.com/gotid/god/api"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
	"time"
)

func main() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Printf("cpu: %d\n", stat.CpuUsage())
		}
	}()

	logx.Disable()
	api.MustNewServer(api.Config{})
}
