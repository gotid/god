package generator

import (
	conf "github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/env"
	"github.com/gotid/god/tools/god/util/console"
	"log"
)

// Generator 定义了 rpc 服务生成所需的环境。
type Generator struct {
	log     console.Console
	cfg     *conf.Config
	verbose bool
}

// NewGenerator 返回 Generator 实例。
func NewGenerator(style string, verbose bool) *Generator {
	cfg, err := conf.NewConfig(style)
	if err != nil {
		log.Fatalln(err)
	}
	return &Generator{
		log:     console.NewColorConsole(verbose),
		cfg:     cfg,
		verbose: verbose,
	}
}

// Prepare 准备用于 rpc 服务生成的环境检测。
func (g *Generator) Prepare() error {
	return env.Prepare(true, true, g.verbose)
}
