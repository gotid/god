package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/api"
	"github.com/gotid/god/api/httpx"
	"net/http"
	"time"
)

var port = flag.Int("port", 3333, "监听端口")

type Request struct {
	User string `form:"user,options=a|b"`
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	deadline, ok := r.Context().Deadline()
	fmt.Println(ok)
	fmt.Printf("%#v\n", deadline)

	var req Request
	err := httpx.Parse(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	time.Sleep(1500 * time.Millisecond)

	httpx.OkJson(w, "hello, "+req.User)
}

func handleApiAbout(w http.ResponseWriter, r *http.Request) {
	deadline, ok := r.Context().Deadline()
	fmt.Println(ok)
	fmt.Printf("%#v\n", deadline)

	var req Request
	err := httpx.Parse(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	time.Sleep(1500 * time.Millisecond)

	httpx.OkJson(w, "hello, "+req.User)
}

func handleApiPrivacy(w http.ResponseWriter, r *http.Request) {
	deadline, ok := r.Context().Deadline()
	fmt.Println(ok)
	fmt.Printf("%#v\n", deadline)

	httpx.OkJson(w, "hello, here's privacy")
}

func main() {
	flag.Parse()

	svr := api.MustNewServer(api.Config{
		Host:    "localhost",
		Port:    *port,
		Timeout: 3000,
	})
	defer svr.Stop()

	svr.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/about",
		Handler: handleAbout,
	}, api.WithTimeout(time.Second))

	svr.AddRoutes([]api.Route{
		{
			Method:  http.MethodGet,
			Path:    "/about",
			Handler: handleApiAbout,
		},
		{
			Method:  http.MethodGet,
			Path:    "/privacy",
			Handler: handleApiPrivacy,
		},
	}, api.WithPrefix("/api"), api.WithTimeout(5*time.Second))

	svr.Start()
}
