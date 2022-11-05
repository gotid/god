package api

import (
	"github.com/gotid/god/tools/god/api/apigen"
	"github.com/gotid/god/tools/god/api/gogen"
	"github.com/gotid/god/tools/god/api/new"
	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:   "api",
		Short: "生成 api 协议模板",
		RunE:  apigen.CreateApiTemplate,
	}

	goCmd = &cobra.Command{
		Use:   "go",
		Short: "根据 api 协议文件，生成 api 示例服务",
		RunE:  gogen.GoCommand,
	}

	newCmd = &cobra.Command{
		Use:   "new",
		Short: "快速生成 api 示例服务",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return new.CreateServiceCommand(args)
		},
	}
)

func init() {
	Cmd.Flags().StringVar(&apigen.VarStringOutput, "o", "", "api 协议文件输出路径")
	Cmd.Flags().StringVar(&apigen.VarStringHome, "home", "", "god 模板主目录，--remote 优先级高于 --home")

	goCmd.Flags().StringVar(&gogen.VarStringDir, "dir", "", "目标目录")
	goCmd.Flags().StringVar(&gogen.VarStringAPI, "api", "", "api 协议文件")
	goCmd.Flags().StringVar(&gogen.VarStringHome, "home", "", "god 模板主目录，--remote 优先级高于 --home")
	goCmd.Flags().StringVar(&gogen.VarStringRemote, "remote", "", "远程 git 模板仓库，优先级高于 home\n\t模板目录要与 https://github.com/gotid/god-template 保持一致")
	goCmd.Flags().StringVar(&gogen.VarStringBranch, "branch", "", "远程仓库分值，与 --remote 配合使用")
	goCmd.Flags().StringVar(&gogen.VarStringStyle, "style", "godesigner", "文件命名样式，详见 [https://github.com/gotid/god/blob/master/tools/god/config/readme.md]")

	newCmd.Flags().StringVar(&gogen.VarStringHome, "home", "", "god 模板主目录，--remote 优先级高于 --home")
	newCmd.Flags().StringVar(&gogen.VarStringRemote, "remote", "", "远程 git 模板仓库，优先级高于 home\n\t模板目录要与 https://github.com/gotid/god-template 保持一致")
	newCmd.Flags().StringVar(&gogen.VarStringBranch, "branch", "", "远程仓库分值，与 --remote 配合使用")
	newCmd.Flags().StringVar(&gogen.VarStringStyle, "style", "godesigner", "文件命名样式，详见 [https://github.com/gotid/god/blob/master/tools/god/config/readme.md]")

	Cmd.AddCommand(goCmd)
	Cmd.AddCommand(newCmd)
}
