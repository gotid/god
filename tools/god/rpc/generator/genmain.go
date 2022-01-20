package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gotid/god/lib/fs"

	"github.com/gotid/god/lib/stringx"
	conf "github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/rpc/parser"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/format"
)

const mainTemplate = `{{.head}}

package main

import (
	"flag"
	"fmt"

	{{.imports}}

	"github.com/gotid/god/lib/conf"
	"github.com/gotid/god/lib/service"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "配置文件")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	srv := server.New{{.serviceNew}}Server(ctx)

	s := rpc.MustNewServer(c.ServerConf, func(grpcServer *grpc.Server) {
		{{.pkg}}.Register{{.service}}Server(grpcServer, srv)

		// 在开发和测试模式下进行服务器反射，可为grpcurl/grpc_cli等提供grpc服务或方法查询。
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
`

func (g *defaultGenerator) GenMain(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	mainFilename, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	fileName := filepath.Join(ctx.GetMain().Filename, fmt.Sprintf("%v.go", mainFilename))
	imports := make([]string, 0)
	pbImport := fmt.Sprintf(`"%v"`, ctx.GetPb().Package)
	svcImport := fmt.Sprintf(`"%v"`, ctx.GetSvc().Package)
	remoteImport := fmt.Sprintf(`"%v"`, ctx.GetServer().Package)
	configImport := fmt.Sprintf(`"%v"`, ctx.GetConfig().Package)
	imports = append(imports, configImport, pbImport, remoteImport, svcImport)
	head := util.GetHead(proto.Name)
	text, err := util.LoadTemplate(category, mainTemplateFile, mainTemplate)
	if err != nil {
		return err
	}

	return util.With("main").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"head":        head,
		"serviceName": strings.ToLower(ctx.GetServiceName().ToCamel()),
		"imports":     strings.Join(imports, fs.NL),
		"pkg":         proto.PbPackage,
		"serviceNew":  stringx.From(proto.Service.Name).ToCamel(),
		"service":     parser.CamelCase(proto.Service.Name),
	}, fileName, false)
}
