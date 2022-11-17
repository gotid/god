package main

import (
	"fmt"
	"github.com/gotid/god/lib/executors"
	"time"
)

// 每90毫秒向批量执行器增加一个任务，批量执行器按10个一批进行执行
func main() {
	be := executors.NewBulkExecutor(func(tasks []interface{}) {
		fmt.Println(len(tasks), tasks)
	}, executors.WithBulkTasks(10))

	for {
		if err := be.Add(1); err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(time.Millisecond * 90)
	}
}
