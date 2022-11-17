package main

import (
	"fmt"
	"github.com/gotid/god/lib/stat"
	"runtime"
	"time"
)

func main() {
	fmt.Println(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU()+10; i++ {
		go func() {
			for {
				select {
				default:
					time.Sleep(time.Microsecond)
				}
			}
		}()
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		usage := stat.CpuUsage()
		fmt.Println("cpu: ", usage)
	}
}
