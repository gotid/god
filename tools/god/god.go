package main

import (
	"github.com/gotid/god/lib/load"
	"github.com/gotid/god/lib/logx"
)

func main() {
	logx.Disable()
	load.Disable()
}
