package cli

import (
	"errors"
	"fmt"
	"github.com/gotid/god/tools/god/rpc/generator"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

var (
	// VarStringOutput 表示输出。
	VarStringOutput string
	// VarStringHome 表示 god home 文件夹。
	VarStringHome string
	// VarStringRemote 表示 god 远程 git 仓库。
	VarStringRemote string
	// VarStringBranch 表示 god 远程 git 分支。
	VarStringBranch string

	// VarStringSliceGoGRPCOpt 表示 grpc 选项。
	VarStringSliceGoGRPCOpt []string
	// VarStringSliceGoOpt 表示 go 选项。
	VarStringSliceGoOpt []string
	// VarStringSliceProtoPath 表示 proto 路径。
	VarStringSliceProtoPath []string
	// VarStringSliceGoOut 表示 go 输出目录。
	VarStringSliceGoOut []string
	// VarStringSliceGoGRPCOut 表示 grpc 输出目录。
	VarStringSliceGoGRPCOut []string
	// VarStringSlicePlugin 表示 grpc 插件。
	VarStringSlicePlugin []string
	// VarStringRPCOut 表示 rpc 代码输出目录。
	VarStringRPCOut string

	// VarStringStyle 表示输出文件的命名风格。
	VarStringStyle string
	// VarBoolVerbose 表示是否输出详情。
	VarBoolVerbose bool
	// VarBoolMultiple 表示是否支持生成多个 rpc 服务。
	VarBoolMultiple bool
)

// RPCTemplate 用于生成 rpc 示例模板。
func RPCTemplate(_ *cobra.Command, _ []string) error {
	protoFile := VarStringOutput
	home := VarStringHome
	remote := VarStringRemote
	branch := VarStringBranch
	if len(remote) > 0 {
		repo, _ := util.CloneIntoGitHome(remote, branch)
		if len(repo) > 0 {
			home = repo
		}
	}
	if len(home) > 0 {
		pathx.RegisterGodHome(home)
	}

	if len(protoFile) == 0 {
		return errors.New("缺少 -o")
	}

	return generator.ProtoTmpl(protoFile)
}

// RPCNew 用于生成 rpc 示例服务。
// 该服务可加速你对 rpc 服务结构的理解。
func RPCNew(_ *cobra.Command, args []string) error {
	rpcName := args[0]
	ext := filepath.Ext(rpcName)
	if len(ext) > 0 {
		return fmt.Errorf("rpc 名称不应设置扩展名：%s", ext)
	}

	home := VarStringHome
	remote := VarStringRemote
	branch := VarStringBranch
	if len(remote) > 0 {
		repo, _ := util.CloneIntoGitHome(remote, branch)
		if len(repo) > 0 {
			home = repo
		}
	}
	if len(home) > 0 {
		pathx.RegisterGodHome(home)
	}

	protoName := rpcName + ".proto"
	filename := filepath.Join(".", rpcName, protoName)
	src, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	err = generator.ProtoTmpl(src)
	if err != nil {
		return err
	}

	ctx := generator.RpcContext{
		Multiple:       false,
		Src:            src,
		IsGooglePlugin: true,
		GoOutput:       filepath.Dir(src),
		GrpcOutput:     filepath.Dir(src),
		Output:         filepath.Dir(src),
		ProtocCmd: fmt.Sprintf("protoc -I=%s %s --go_out=%s --go-grpc_out=%s",
			filepath.Dir(src), filepath.Base(src), filepath.Dir(src), filepath.Dir(src)),
	}

	grpcOptList := VarStringSliceGoGRPCOpt
	if len(grpcOptList) > 0 {
		ctx.ProtocCmd += " --go-grpc_opt=" + strings.Join(grpcOptList, ",")
	}

	goOptList := VarStringSliceGoOpt
	if len(goOptList) > 0 {
		ctx.ProtocCmd += " --go_opt=" + strings.Join(goOptList, ",")
	}

	style := VarStringStyle
	verbose := VarBoolVerbose

	g := generator.NewGenerator(style, verbose)
	return g.Generate(&ctx)
}
