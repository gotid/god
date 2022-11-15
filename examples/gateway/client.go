package main

import (
	"context"
	"fmt"
	"github.com/gotid/god/api/httpc"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for t := range ticker.C {
		resp, err := httpc.Do(context.Background(), http.MethodGet, "http://localhost:8888/pingHello/"+t.Format("2006-01-02 15:04:05"), nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		io.Copy(os.Stdout, resp.Body)
	}
}
