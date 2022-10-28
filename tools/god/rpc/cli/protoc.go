package cli

import (
	"errors"
	"github.com/gotid/god/tools/god/rpc/generator"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	errInvalidGrpcOutput = errors.New("RPC：缺失 --go-grpc_out")
	errInvalidGoOutput   = errors.New("RPC：缺失 --go_out")
	errInvalidRpcOutput  = errors.New("RPC：缺失 rpc 输出目录，请用 --rpc_out 指定输出目录")
)

// RPCProtoc 通过 protoc 生成 grpc 代码，通过 god 生成 rpc 代码。
func RPCProtoc(_ *cobra.Command, args []string) error {
	protocArgs := wrapProtocCmd("protoc", args)
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	source := args[0]
	grpcOutList := VarStringSliceGoGRPCOut
	goOutList := VarStringSliceGoOut
	rpcOut := VarStringRPCOut
	style := VarStringStyle
	home := VarStringHome
	remote := VarStringRemote
	branch := VarStringBranch
	verbose := VarBoolVerbose

	if len(grpcOutList) == 0 {
		return errInvalidGrpcOutput
	}
	if len(goOutList) == 0 {
		return errInvalidGoOutput
	}

	goOut := goOutList[len(goOutList)-1]
	grpcOut := grpcOutList[len(grpcOutList)-1]
	if len(goOut) == 0 {
		return errInvalidGrpcOutput
	}
	if len(grpcOut) == 0 {
		return errInvalidRpcOutput
	}
	goOutAbs, err := filepath.Abs(goOut)
	if err != nil {
		return err
	}
	grpcOutAbs, err := filepath.Abs(grpcOut)
	if err != nil {
		return err
	}
	err = pathx.MkdirIfNotExist(goOutAbs)
	if err != nil {
		return err
	}
	err = pathx.MkdirIfNotExist(grpcOutAbs)
	if err != nil {
		return err
	}

	if len(remote) > 0 {
		repo, _ := util.CloneIntoGitHome(remote, branch)
		if len(repo) > 0 {
			home = repo
		}
	}
	if len(home) > 0 {
		pathx.RegisterGodHome(home)
	}

	if !filepath.IsAbs(rpcOut) {
		rpcOut = filepath.Join(pwd, rpcOut)
	}

	isGooglePlugin := len(grpcOut) > 0
	goOut, err = filepath.Abs(goOut)
	if err != nil {
		return err
	}
	grpcOut, err = filepath.Abs(grpcOut)
	if err != nil {
		return err
	}
	rpcOut, err = filepath.Abs(rpcOut)
	if err != nil {
		return err
	}

	ctx := generator.RpcContext{
		Multiple:       VarBoolMultiple,
		Src:            source,
		GoOutput:       goOut,
		GrpcOutput:     grpcOut,
		IsGooglePlugin: isGooglePlugin,
		Output:         rpcOut,
		ProtocCmd:      strings.Join(protocArgs, " "),
	}
	g := generator.NewGenerator(style, verbose)
	return g.Generate(&ctx)
}

func wrapProtocCmd(name string, args []string) []string {
	ret := append([]string{name}, args...)
	for _, path := range VarStringSliceProtoPath {
		ret = append(ret, "--proto_path", path)
	}
	for _, path := range VarStringSliceGoOpt {
		ret = append(ret, "--go_opt", path)
	}
	for _, path := range VarStringSliceGoGRPCOpt {
		ret = append(ret, "--go-grpc_opt", path)
	}
	for _, path := range VarStringSliceGoOut {
		ret = append(ret, "--go_out", path)
	}
	for _, path := range VarStringSliceGoGRPCOut {
		ret = append(ret, "--go-grpc_out", path)
	}
	for _, plugin := range VarStringSlicePlugin {
		ret = append(ret, "--plugin="+plugin)
	}

	return ret
}
