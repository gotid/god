package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/lib/proc"
	"log"
	"runtime"
	"strconv"
	"time"
)

const numItems = 1000000

var round = flag.Int("r", 3, "回合数")

func main() {
	defer proc.StartProfile().Stop()

	flag.Parse()

	fmt.Println(getMemUsage())

	for i := 0; i < *round; i++ {
		do()
	}

}

func do() {
	tw, err := collection.NewTimingWheel(time.Second, 100, execute)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < numItems; i++ {
		key := strconv.Itoa(i)
		tw.SetTimer(key, key, time.Second*5)
	}

	fmt.Println(getMemUsage())

}

func execute(key any, val any) {

}

func getMemUsage() string {
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("堆分配 = %dMiB", toMib(m.HeapAlloc))
}

func toMib(b uint64) uint64 {
	return b / 1024 / 1024
}
