package env

import (
	"fmt"
	"github.com/gotid/god/tools/god/pkg/env"
	"github.com/gotid/god/tools/god/pkg/protoc"
	"github.com/gotid/god/tools/god/pkg/protocgengo"
	"github.com/gotid/god/tools/god/pkg/protocgengogrpc"
	"github.com/gotid/god/tools/god/util/console"
	"strings"
	"time"
)

type bin struct {
	name   string
	exists bool
	get    func(cacheDir string) (string, error)
}

var bins = []bin{
	{
		name:   "protoc",
		exists: protoc.Exists(),
		get:    protoc.Install,
	},
	{
		name:   "protoc-gen-go",
		exists: protocgengo.Exists(),
		get:    protocgengo.Install,
	},
	{
		name:   "protoc-gen-go-grpc",
		exists: protocgengogrpc.Exists(),
		get:    protocgengogrpc.Install,
	},
}

func Prepare(install, force, verbose bool) error {
	log := console.NewColorConsole(verbose)
	pending := true
	log.Info("[god-env]：准备检查环境")

	defer func() {
		if p := recover(); p != nil {
			log.Error("%+v", p)
			return
		}
		if pending {
			log.Success("\n[god-env]：恭喜！你的 god 环境已就绪！")
		} else {
			log.Error(`
[god-env]：环境检查已完成，某些依赖在系统路径中未找到，你可执行命令 'god env check --install' 进行安装。
详情参考 'god env check --help'`)
		}
	}()

	for _, e := range bins {
		time.Sleep(200 * time.Millisecond)
		log.Info("")
		log.Info("[god-env]：查找 %q", e.name)
		if e.exists {
			log.Success("[god-env]：%q 已安装", e.name)
			continue
		}
		log.Warning("[god-env]：%q 在环境变量中未找到", e.name)
		if install {
			installBin := func() {
				log.Info("[god-env]：准备安装 %q", e.name)
				path, err := e.get(env.Get(env.GodCache))
				if err != nil {
					log.Error("[god-env]：安装错误：%+v", err)
					pending = false
				} else {
					log.Success("[god-env]：%q 已安装在 %q", e.name, path)
				}
			}
			if force {
				installBin()
				continue
			}
			console.Info("[god-env]：你要安装 %q 吗？[y: 安装，n：不安装]", e.name)
			for {
				var in string
				fmt.Scanln(&in)
				var brk bool
				switch {
				case strings.EqualFold(in, "y"):
					installBin()
					brk = true
				case strings.EqualFold(in, "n"):
					pending = false
					console.Info("[god-env]：%q 安装已取消", e.name)
					brk = true
				default:
					console.Error("[god-env]：无效输入，请输入 'y' 表示同意安装，'n' 表示取消")
				}
				if brk {
					break
				}
			}
		} else {
			pending = false
		}
	}

	return nil
}
