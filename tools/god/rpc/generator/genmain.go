package generator

import (
	_ "embed"
	"fmt"
	conf "github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/rpc/parser"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/format"
	"github.com/gotid/god/tools/god/util/pathx"
	"path/filepath"
	"strings"
)

//go:embed main.tpl
var mainTemplate string

type MainServiceTemplateData struct {
	Service   string
	ServerPkg string
	Pkg       string
}

// GenMain 生成 rpc 服务主文件，作为程序的调用入口。
func (g *Generator) GenMain(ctx DirContext, proto parser.Proto, cfg *conf.Config, c *RpcContext) error {
	mainFilename, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	filename := filepath.Join(ctx.GetMain().Filename, fmt.Sprintf("%s.go", mainFilename))
	imports := make([]string, 0)
	pbImport := fmt.Sprintf(`"%v"`, ctx.GetPb().Package)
	svcImport := fmt.Sprintf(`"%v"`, ctx.GetSvc().Package)
	configImport := fmt.Sprintf(`"%v"`, ctx.GetConfig().Package)
	imports = append(imports, configImport, pbImport, svcImport)

	var serviceNames []MainServiceTemplateData
	for _, e := range proto.Service {
		var (
			remoteImport string
			serverPkg    string
		)

		if !c.Multiple {
			serverPkg = "server"
			remoteImport = fmt.Sprintf(`"%v"`, ctx.GetServer().Package)
		} else {
			childPkg, err := ctx.GetServer().GetChildPackage(e.Name)
			if err != nil {
				return err
			}

			serverPkg = filepath.Base(childPkg + "Server")
			remoteImport = fmt.Sprintf(`%s "%v"`, serverPkg, childPkg)
		}

		imports = append(imports, remoteImport)
		serviceNames = append(serviceNames, MainServiceTemplateData{
			Service:   parser.CamelCase(e.Name),
			ServerPkg: serverPkg,
			Pkg:       proto.PbPackage,
		})
	}

	text, err := pathx.LoadTemplate(category, mainTemplateFile, mainTemplate)
	if err != nil {
		return err
	}

	serviceName, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	return util.With("main").GoFmt(true).Parse(text).SaveTo(map[string]any{
		"serviceName":  serviceName,
		"imports":      strings.Join(imports, pathx.NL),
		"pkg":          proto.PbPackage,
		"serviceNames": serviceNames,
	}, filename, false)
}
