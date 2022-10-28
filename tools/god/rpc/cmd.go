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

	protocCmd = &cobra.Command{
		Use:     "protoc",
		Short:   "生成 grpc 代码",
		Example: "god rpc protoc xxx.proto --go_out=./pb --go-grpc_out=./pb --rpc_out=.",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE:    cli.RPCProtoc,
	}
)

func init() {
	Cmd.Flags().StringVar(&cli.VarStringOutput, "o", "", "输出一个示例 proto 文件")
	Cmd.Flags().StringVar(&cli.VarStringHome, "home", "", "god 模板主目录，--remote 优先级高于 --home")

	newCmd.Flags().StringVar(&cli.VarStringStyle, "style", "godesigner", "文件命名样式")
	newCmd.Flags().BoolVarP(&cli.VarBoolVerbose, "verbose", "v", false, "启用日志输出")

	protocCmd.Flags().BoolVarP(&cli.VarBoolMultiple, "multiple", "m", false, "生成多个rpc服务")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSliceGoOut, "go_out", nil, "go 输出目录")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSliceGoGRPCOut, "go-grpc_out", nil, "grpc 输出目录")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSliceGoOpt, "go_opt", nil, "go 选项")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSliceGoGRPCOpt, "go-grpc_opt", nil, "grpc 选项")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSlicePlugin, "plugin", nil, "grpc 插件")
	protocCmd.Flags().StringSliceVarP(&cli.VarStringSliceProtoPath, "proto_path", "I", nil, "proto 路径")
	protocCmd.Flags().StringVar(&cli.VarStringRPCOut, "rpc_out", "", "rpc 代码输出目录")
	protocCmd.Flags().StringVar(&cli.VarStringStyle, "style", "godesigner", "文件命名样式，详见 [https://github.com/gotid/god/tree/master/tools/god/config/readme.md]")
	protocCmd.Flags().StringVar(&cli.VarStringHome, "home", "", "The goctl home "+
		"path of the template, --home and --remote cannot be set at the same time, if they are, "+
		"--remote has higher priority")
	protocCmd.Flags().StringVar(&cli.VarStringRemote, "remote", "", "The remote "+
		"git repo of the template, --home and --remote cannot be set at the same time, if they are, "+
		"--remote has higher priority\n\tThe git repo directory must be consistent with the "+
		"https://github.com/zeromicro/go-zero-template directory structure")
	protocCmd.Flags().StringVar(&cli.VarStringBranch, "branch", "",
		"The branch of the remote repo, it does work with --remote")
	protocCmd.Flags().BoolVarP(&cli.VarBoolVerbose, "verbose", "v", false, "Enable log output")
	protocCmd.Flags().MarkHidden("go_out")
	protocCmd.Flags().MarkHidden("go-grpc_out")
	protocCmd.Flags().MarkHidden("go_opt")
	protocCmd.Flags().MarkHidden("go-grpc_opt")
	protocCmd.Flags().MarkHidden("plugin")
	protocCmd.Flags().MarkHidden("proto_path")

	Cmd.AddCommand(newCmd)
}
