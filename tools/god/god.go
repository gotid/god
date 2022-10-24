package main

import (
	"github.com/gotid/god/lib/load"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/tools/god/cmd"
)

func main() {
	logx.Disable()
	load.Disable()
	cmd.Execute()
}
