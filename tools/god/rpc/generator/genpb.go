package generator

import (
	"fmt"
	"github.com/gotid/god/tools/god/rpc/execx"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// GenPb 生成 pb.go 文件，是用于 protoc 生成 grpc 的封装层。
// 目前 god 代码生成器尚未完成集成 protoc 的命令和标识符，目前已支持 proto_path(-I)。
func (g *Generator) GenPb(ctx DirContext, c *RpcContext) error {
	return g.genPbDirect(ctx, c)
}

func (g *Generator) genPbDirect(ctx DirContext, c *RpcContext) error {
	g.log.Debug("[command]: %s", c.ProtocCmd)
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	_, err = execx.Run(c.ProtocCmd, pwd)
	if err != nil {
		return err
	}

	return g.setPbDir(ctx, c)
}

func (g *Generator) setPbDir(ctx DirContext, c *RpcContext) error {
	pbDir, err := findPbFile(c.GoOutput, false)
	if err != nil {
		return err
	}
	if len(pbDir) == 0 {
		return fmt.Errorf("pb.go 未出现在目录 %q 中", c.GoOutput)
	}

	grpcDir, err := findPbFile(c.GrpcOutput, true)
	if err != nil {
		return err
	}
	if len(grpcDir) == 0 {
		return fmt.Errorf("_grpc.pb.go 未出现在目录 %q 中", c.GrpcOutput)
	}

	if pbDir != grpcDir {
		return fmt.Errorf("pb.go 和 _grpc.pb.go 必须位于相同目录：\n pb.go: %s\n _grpc.pb.go: %s", pbDir, grpcDir)
	}
	if pbDir == c.Output {
		return fmt.Errorf("pb.go 和 _grpc.pb.go 的输出目录不能相同：\n pb 输出目录：%s\n rpc 输出目录：%s", pbDir, c.Output)
	}

	ctx.SetPbDir(pbDir, grpcDir)

	return nil
}

const (
	pbSuffix   = "pb.go"
	grpcSuffix = "_grpc.pb.go"
)

func findPbFile(current string, grpc bool) (string, error) {
	fileSystem := os.DirFS(current)
	var ret string
	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, pbSuffix) {
			if grpc {
				if strings.HasSuffix(path, grpcSuffix) {
					ret = path
					return os.ErrExist
				}
			} else if !strings.HasSuffix(path, grpcSuffix) {
				ret = path
				return os.ErrExist
			}
		}
		return nil
	})
	if err == os.ErrExist {
		return filepath.Dir(filepath.Join(current, ret)), nil
	}

	return "", err
}
