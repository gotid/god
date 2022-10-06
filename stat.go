package main

import (
	"fmt"
	"github.com/gotid/god/lib/stat"
)

func main() {
	cpuUsage := stat.CpuUsage()
	fmt.Printf("cpu: %d", cpuUsage)

	select {}
}
