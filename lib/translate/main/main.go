package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/gotid/god/lib/translate"
	"github.com/urfave/cli"
)

var commands = []cli.Command{
	{
		Name:      "baidu",
		ShortName: "bd",
		Usage:     "百度翻译",
		Action:    baiduTranslate,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "query, q",
				Usage: "待翻译文本",
			},
			cli.StringFlag{
				Name:  "time, t",
				Usage: "是否计时",
			},
		},
	},
}

func main() {
	app := cli.NewApp()
	app.Usage = "调用百度、谷歌等翻译器"
	app.Version = fmt.Sprintf("%s %s %s",
		time.Now().Format("2006-01-02"),
		runtime.GOOS,
		runtime.GOARCH,
	)
	app.Commands = commands
	if err := app.Run(os.Args); err != nil {
		fmt.Println("启动错误", err)
	}
}

func baiduTranslate(ctx *cli.Context) error {
	query := ctx.String("query")
	translate.Baidu.Time(ctx.Bool("time"))
	fmt.Println(query, "->", translate.Baidu.ToZh(query))
	return nil
}
