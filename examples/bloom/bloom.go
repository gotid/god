package main

import (
	"encoding/binary"
	"fmt"
	bloomv3 "github.com/bits-and-blooms/bloom/v3"
	"github.com/gotid/god/lib/bloom"
	"github.com/gotid/god/lib/store/redis"
)

func main() {
	total := 10000

	store := redis.New("localhost:6379")
	filter := bloom.New(store, "test:bloom", uint(total))
	for i := 0; i < total; i++ {
		i := uint32(i)
		n1 := make([]byte, 4)
		binary.BigEndian.PutUint32(n1, i)
		filter.Add(n1)
	}

	count := 0
	for i := 0; i < total+1000; i++ {
		i := uint32(i)
		n1 := make([]byte, 4)
		binary.BigEndian.PutUint32(n1, i)
		if exists, _ := filter.Exists(n1); exists {
			count++
		}
	}
	fmt.Println("redis   已匹配的数量", count)

	filterv3 := bloomv3.NewWithEstimates(uint(total), 0.01)

	for i := 0; i < total; i++ {
		i := uint32(i)
		n1 := make([]byte, 4)
		binary.BigEndian.PutUint32(n1, i)
		filterv3.Add(n1)
	}

	count2 := 0
	for i := 0; i < total+1000; i++ {
		i := uint32(i)
		n1 := make([]byte, 4)
		binary.BigEndian.PutUint32(n1, i)
		if filterv3.Test(n1) {
			count2++
		}
	}
	fmt.Println("bloomv3 已匹配的数量", count2)
}
