package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"git.zc0901.com/go/god/lib/stat"

	"git.zc0901.com/go/god/lib/load"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/tools/god/api/apigen"
	"git.zc0901.com/go/god/tools/god/api/docgen"
	"git.zc0901.com/go/god/tools/god/api/format"
	"git.zc0901.com/go/god/tools/god/api/gogen"
	"git.zc0901.com/go/god/tools/god/api/new"
	"git.zc0901.com/go/god/tools/god/api/validate"
	"git.zc0901.com/go/god/tools/god/mysql/command"
	rpc "git.zc0901.com/go/god/tools/god/rpc/cli"
	"github.com/urfave/cli"
)

var (
	BuildTime = time.Now().Format("20060102")
	commands  = []cli.Command{
		{
			Name:   "mysql",
			Usage:  "从数据源生成 MySQL 模型层代码",
			Action: command.GenCodeFromDSN,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dsn",
					Usage: `数据库连接地址，如 "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label"`,
				},
				cli.StringFlag{
					Name:  "table, t",
					Usage: `表名，多表以英文逗号分隔，如 "node,tag,channel"`,
				},
				cli.StringFlag{
					Name:  "dir, d",
					Usage: "目标文件夹",
				},
				cli.BoolFlag{
					Name:  "cache, c",
					Usage: "生成带缓存的数据访问层[可选]",
				},
			},
		},
		{
			Name:  "rpc",
			Usage: "生成 GRPC 代码模板",
			Subcommands: []cli.Command{
				{
					Name:  "new",
					Usage: `生成 RPC 层示例服务`,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:     "style",
							Required: false,
							Usage:    "文件命名风格，详见 [https://github.com/tal-tech/go-zero/tree/master/tools/goctl/config/readme.md]",
						},
						cli.BoolFlag{
							Name:  "idea",
							Usage: "命令行执行环境是否为 Idea 插件。[可选]",
						},
					},
					Action: rpc.RpcNew,
				},
				{
					Name:  "template",
					Usage: `生成 proto 模板`,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "out, o",
							Usage: "proto 协议目标路径。",
						},
					},
					Action: rpc.RpcTemplate,
				},
				{
					Name:  "proto",
					Usage: `从 proto 协议生成 RPC 模板`,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "src, s",
							Usage: "proto 协议文件路径",
						},
						cli.StringSliceFlag{
							Name:  "proto_path, I",
							Usage: `protoc 原始命令路径，指定搜索导包路径。[导包]`,
						},
						cli.StringFlag{
							Name:  "dir, d",
							Usage: `代码目标路径`,
						},
						cli.StringFlag{
							Name:     "style",
							Required: false,
							Usage:    "文件命名风格，详见 [https://github.com/tal-tech/go-zero/tree/master/tools/goctl/config/readme.md]",
						},
						cli.BoolFlag{
							Name:  "idea",
							Usage: "命令行执行环境是否为 Idea 插件。[可选]",
						},
					},
					Action: rpc.Rpc,
				},
			},
		},
		{
			Name:  "api",
			Usage: "生成 API 层服务模板",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "o",
					Usage: "输出的 API 文件路径",
				},
			},
			Action: apigen.ApiCommand,
			Subcommands: []cli.Command{
				{
					Name:   "new",
					Usage:  "快速生成 API 服务模板",
					Action: new.NewService,
				},
				{
					Name:  "format",
					Usage: "格式化 API 文件",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "dir",
							Usage: "待格式化目录",
						},
						cli.BoolFlag{
							Name:     "iu",
							Usage:    "是否忽略更新",
							Required: false,
						},
						cli.BoolFlag{
							Name:     "stdin",
							Usage:    "use stdin to input api doc content, press \"ctrl + d\" to send EOF",
							Required: false,
						},
					},
					Action: format.GoFormatApi,
				},
				{
					Name:  "validate",
					Usage: "验证 API 文件",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "api",
							Usage: "待验证 API 文件",
						},
					},
					Action: validate.GoValidateApi,
				},
				{
					Name:  "doc",
					Usage: "生成文档文件",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "dir",
							Usage: "保存文件夹",
						},
					},
					Action: docgen.DocCommand,
				},
				{
					Name:  "go",
					Usage: "生成 Go 版本 API 服务",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "dir",
							Usage: "目标目录",
						},
						cli.StringFlag{
							Name:  "api",
							Usage: "指定的 API 文件",
						},
						cli.StringFlag{
							Name:     "style",
							Required: false,
							Usage:    "文件命名风格，详见 [https://github.com/tal-tech/go-zero/tree/master/tools/goctl/config/readme.md]",
						},
					},
					Action: gogen.GoCommand,
				},
			},
		},
	}
)

func main() {
	logx.Disable()
	load.Disable()
	stat.CpuUsage()

	app := cli.NewApp()
	app.Usage = "God 代码生成器"
	app.Version = fmt.Sprintf("%s %s/%s", BuildTime, runtime.GOOS, runtime.GOARCH)
	app.Commands = commands
	if err := app.Run(os.Args); err != nil {
		fmt.Println("错误：", err)
	}
}
