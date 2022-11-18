package rpc

import (
	"github.com/gotid/god/tools/god/rpc/cli"
	"github.com/spf13/cobra"
)

var (
	// Cmd 描述了一个 rpc 命令。
	Cmd = &cobra.Command{
		Use:   "rpc",
		Short: "生成 proto 协议模板",
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
		Short:   "根据 proto 文件，生成 rpc 示例服务",
		Example: "god rpc protoc xxx.proto --go_out=./pb --go-grpc_out=./pb --rpc_out=.",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE:    cli.RPCProtoc,
	}
)

func init() {
	Cmd.Flags().StringVar(&cli.VarStringOutput, "o", "", "输出一个 proto 协议模板")
	Cmd.Flags().StringVar(&cli.VarStringHome, "home", "", "god 模板主目录，--remote 优先级高于 --home")

	protocCmd.Flags().BoolVarP(&cli.VarBoolMultiple, "multiple", "m", false, "生成多个rpc服务")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSliceGoOut, "go_out", []string{"./"}, "go 输出目录")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSliceGoOpt, "go_opt", nil, "go 选项")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSliceGoGrpcOut, "go-grpc_out", []string{"./"}, "grpc 输出目录")
	protocCmd.Flags().StringVar(&cli.VarStringRPCOut, "rpc_out", "", "rpc 代码输出目录")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSliceGoGrpcOpt, "go-grpc_opt", nil, "grpc 选项")
	protocCmd.Flags().StringSliceVar(&cli.VarStringSlicePlugin, "plugin", nil, "grpc 插件")
	protocCmd.Flags().StringSliceVarP(&cli.VarStringSliceProtoPath, "proto_path", "I", nil, "proto 路径")
	protocCmd.Flags().StringVar(&cli.VarStringStyle, "style", "godesigner", "文件命名样式，详见 [https://github.com/gotid/god/blob/master/tools/god/config/readme.md]")
	protocCmd.Flags().StringVar(&cli.VarStringHome, "home", "", "god 模板主目录，--remote 优先级高于 --home")
	protocCmd.Flags().StringVar(&cli.VarStringRemote, "remote", "", "远程 git 模板仓库，优先级高于 home\n\t模板目录要与 https://github.com/gotid/god-template 保持一致")
	protocCmd.Flags().StringVar(&cli.VarStringBranch, "branch", "", "远程仓库分值，与 --remote 配合使用")
	protocCmd.Flags().BoolVarP(&cli.VarBoolVerbose, "verbose", "v", false, "启用日志输出")
	_ = protocCmd.Flags().MarkHidden("go_out")
	_ = protocCmd.Flags().MarkHidden("go-grpc_out")
	_ = protocCmd.Flags().MarkHidden("go_opt")
	_ = protocCmd.Flags().MarkHidden("go-grpc_opt")
	_ = protocCmd.Flags().MarkHidden("plugin")
	_ = protocCmd.Flags().MarkHidden("proto_path")

	newCmd.Flags().StringVar(&cli.VarStringStyle, "style", "godesigner", "文件命名样式")
	newCmd.Flags().BoolVarP(&cli.VarBoolVerbose, "verbose", "v", false, "启用日志输出")

	Cmd.AddCommand(protocCmd)
	Cmd.AddCommand(newCmd)
}
