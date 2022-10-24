package generator

import (
	"fmt"
	"github.com/gotid/god/tools/god/rpc/parser"
	"github.com/gotid/god/tools/god/util/console"
	"github.com/gotid/god/tools/god/util/ctx"
	"github.com/gotid/god/tools/god/util/pathx"
	"path/filepath"
)

type RpcContext struct {
	// proto 协议的源文件。
	Src string
	// 用于生成 proto 文件的命令。
	ProtocCmd string
	// 用于存储生成后的 proto 文件的目录。
	ProtoGenGrpcDir string
	// 用于存储生成后的 go 文件的目录。
	ProtoGenGoDir string
	// 指示 proto 文件是否由谷歌插件生成的标志位。
	IsGooglePlugin bool
	// 是生成后的 go 文件的输出目录。
	GoOutput string
	// 是生成后的 grpc 文件的输出目录。
	GrpcOutput string
	// 是生成后的文件的输出目录。
	Output string
	// 指示 proto 文件是否在 multiple 模式下生成的标志位。
	Multiple bool
}

// Generate 通过 proto 文件、代码存储目录和导入参数生成一个 rpc 服务。
func (g *Generator) Generate(rpcCtx *RpcContext) error {
	abs, err := filepath.Abs(rpcCtx.Output)
	if err != nil {
		return err
	}

	err = pathx.MkdirIfNotExist(abs)
	if err != nil {
		return err
	}

	err = g.Prepare()
	if err != nil {
		return err
	}

	projectCtx, err := ctx.Prepare(abs)
	if err != nil {
		return err
	}

	// 解析 proto 文件
	p := parser.NewDefaultProtoParser()
	proto, err := p.Parse(rpcCtx.Src, rpcCtx.Multiple)
	if err != nil {
		return err
	}

	// 建立项目子目录
	dirCtx, err := mkdir(projectCtx, proto, g.cfg, rpcCtx)
	if err != nil {
		return err
	}
	fmt.Println(dirCtx)

	// 生成 etc 配置文件
	//err = g.GenEtc(dirCtx, proto, g.cfg)
	//if err != nil {
	//	return err
	//}

	// 生成 Pb

	// 生成 config 配置代码

	// 生成服务

	// 生成逻辑

	// 生成服务器

	// 生成主入口

	// 生成调用

	console.NewColorConsole().MarkDone()

	return err
}
