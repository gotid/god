package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/template"

	"github.com/gotid/god/tools/god/internal/version"
	"github.com/gotid/god/tools/god/model"
	"github.com/gotid/god/tools/god/rpc"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/withfig/autocomplete-tools/integrations/cobra"
)

const (
	dash       = "-"
	doubleDash = "--"
	assign     = "="
)

var (
	//go:embed usage.tpl
	usageTpl string

	rootCmd = &cobra.Command{
		Use:   "god",
		Short: "god 代码生成器",
		Long:  "用于生成api接口、rpc服务、model模型代码",
	}
)

// Execute 执行给定命令。
func Execute() {
	os.Args = supportGoStdFlag(os.Args)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(aurora.Red(err.Error()))
		os.Exit(1)
	}
}

func supportGoStdFlag(args []string) []string {
	copyArgs := append([]string(nil), args...)
	arg := args[:1]
	parentCmd, _, err := rootCmd.Traverse(arg)
	if err != nil {
		return copyArgs
	}

	for idx, arg := range copyArgs[0:] {
		parentCmd, _, err = parentCmd.Traverse([]string{arg})
		if err != nil { // 忽略，让 cobra 处理该错误。
			break
		}
		if !strings.HasPrefix(arg, dash) {
			continue
		}

		flagExpr := strings.TrimPrefix(arg, doubleDash)
		flagExpr = strings.TrimPrefix(flagExpr, dash)
		flagName, flagValue := flagExpr, ""
		assignIndex := strings.Index(flagExpr, assign)
		if assignIndex > 0 {
			flagName = flagExpr[:assignIndex]
			flagValue = flagExpr[assignIndex:]
		}

		if !isBuiltin(flagName) {
			// 仅处理用户自定义标识符
			f := parentCmd.Flag(flagName)
			if f == nil {
				continue
			}
			if f.Shorthand == flagName {
				continue
			}
		}

		goStyleFlag := doubleDash + flagName
		if assignIndex > 0 {
			goStyleFlag += flagValue
		}

		copyArgs[idx] = goStyleFlag
	}

	return copyArgs
}

func isBuiltin(name string) bool {
	return name == "version" || name == "help"
}

func init() {
	cobra.AddTemplateFuncs(template.FuncMap{
		"blue":    blue,
		"green":   green,
		"rPadX":   rPadX,
		"rainbow": rainbow,
	})

	rootCmd.Version = fmt.Sprintf(
		"%s %s/%s",
		version.BuildVersion, runtime.GOOS, runtime.GOARCH)

	rootCmd.SetUsageTemplate(usageTpl)
	rootCmd.AddCommand(rpc.Cmd)
	rootCmd.AddCommand(model.Cmd)
	rootCmd.AddCommand(cobracompletefig.CreateCompletionSpecCommand())
}
