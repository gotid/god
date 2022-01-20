package main

import (
	"time"

	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/lib/logx"
)

type TimeHolder struct {
	Date time.Time `json:"date"`
}

func main() {
	th := &TimeHolder{}
	err := conf.LoadConfig("./date.yml", th)
	if err != nil {
		logx.Error(err)
	}
	logx.Infof("%+v", th)
}
