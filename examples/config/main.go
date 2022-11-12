package main

import (
	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/lib/logx"
)

type TimeHolder struct {
	Date string `json:"date"`
}

func main() {
	th := &TimeHolder{}
	err := conf.Load("/Users/zs/Github/gotid/god/examples/config/data.yml", th)
	if err != nil {
		logx.Error(err)
	}
	logx.Infof("%+v", th)
}
