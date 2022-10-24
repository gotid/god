package rpc

import (
	"github.com/gotid/god/tools/god/rpc/cli"
	"github.com/spf13/cobra"
)

var (
	// Cmd 描述了一个 rpc 命令。
	Cmd = &cobra.Command{
		Use:   "rpc",
		Short: "生成 rpc 代码",
		RunE:  cli.RPCTemplate,
	}

	newCmd = &cobra.Command{
		Use:   "new",
		Short: "生成 rpc 示例服务",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE:  cli.RPCNew,
	}
)

func init() {
	Cmd.Flags().StringVar(&cli.VarStringOutput, "o", "", "输出一个示例 proto 文件")
	Cmd.Flags().StringVar(&cli.VarStringHome, "home", "", "模板的 god home 路径，--remote 优先级高于 --home")

	Cmd.AddCommand(newCmd)
}
