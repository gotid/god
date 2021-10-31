package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"git.zc0901.com/go/god/lib/fs"

	"git.zc0901.com/go/god/lib/collection"
	"git.zc0901.com/go/god/lib/stringx"
	conf "git.zc0901.com/go/god/tools/god/config"
	"git.zc0901.com/go/god/tools/god/rpc/parser"
	"git.zc0901.com/go/god/tools/god/util"
	"git.zc0901.com/go/god/tools/god/util/format"
)

const (
	callTemplateText = `{{.head}}

//go:generate mockgen -destination ./{{.name}}_mock.go -package {{.filePackage}} -source $GOFILE

package {{.filePackage}}

import (
	"pathvar"

	{{.package}}

	"git.zc0901.com/go/god/rpc"
)

type (
	{{.alias}}

	{{.serviceName}} interface {
		{{.interface}}
	}

	default{{.serviceName}} struct {
		cli rpc.Client
	}
)

func New{{.serviceName}}(cli rpc.Client) {{.serviceName}} {
	return &default{{.serviceName}}{
		cli: cli,
	}
}

{{.functions}}
`

	callInterfaceFunctionTemplate = `{{if .hasComment}}{{.comment}}
{{end}}{{.method}}(ctx pathvar.Context,req *{{.pbRequest}}) (*{{.pbResponse}},error)`

	callFunctionTemplate = `
{{if .hasComment}}{{.comment}}{{end}}
func (m *default{{.serviceName}}) {{.method}}(ctx pathvar.Context,req *{{.pbRequest}}) (*{{.pbResponse}}, error) {
	client := {{.package}}.New{{.rpcServiceName}}Client(m.cli.Conn())
	return client.{{.method}}(ctx, req)
}
`
)

func (g *defaultGenerator) GenCall(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetCall()
	service := proto.Service
	head := util.GetHead(proto.Name)

	callFilename, err := format.FileNamingFormat(cfg.NamingFormat, service.Name)
	if err != nil {
		return err
	}

	filename := filepath.Join(dir.Filename, fmt.Sprintf("%s.go", callFilename))
	functions, err := g.genFunction(proto.PbPackage, service)
	if err != nil {
		return err
	}

	iFunctions, err := g.getInterfaceFuncs(service)
	if err != nil {
		return err
	}

	text, err := util.LoadTemplate(category, callTemplateFile, callTemplateText)
	if err != nil {
		return err
	}

	alias := collection.NewSet()
	for _, item := range proto.Message {
		alias.AddStr(fmt.Sprintf("%s = %s", parser.CamelCase(item.Name), fmt.Sprintf("%s.%s", proto.PbPackage, parser.CamelCase(item.Name))))
	}

	err = util.With("shared").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"name":        callFilename,
		"alias":       strings.Join(alias.KeysStr(), fs.NL),
		"head":        head,
		"filePackage": dir.Base,
		"package":     fmt.Sprintf(`"%s"`, ctx.GetPb().Package),
		"serviceName": stringx.From(service.Name).ToCamel(),
		"functions":   strings.Join(functions, fs.NL),
		"interface":   strings.Join(iFunctions, fs.NL),
	}, filename, true)
	return err
}

func (g *defaultGenerator) genFunction(goPackage string, service parser.Service) ([]string, error) {
	functions := make([]string, 0)
	for _, rpc := range service.RPC {
		text, err := util.LoadTemplate(category, callFunctionTemplateFile, callFunctionTemplate)
		if err != nil {
			return nil, err
		}

		comment := parser.GetComment(rpc.Doc())
		buffer, err := util.With("sharedFn").Parse(text).Execute(map[string]interface{}{
			"serviceName":    stringx.From(service.Name).ToCamel(),
			"rpcServiceName": parser.CamelCase(service.Name),
			"method":         parser.CamelCase(rpc.Name),
			"package":        goPackage,
			"pbRequest":      parser.CamelCase(rpc.RequestType),
			"pbResponse":     parser.CamelCase(rpc.ReturnsType),
			"hasComment":     len(comment) > 0,
			"comment":        comment,
		})
		if err != nil {
			return nil, err
		}

		functions = append(functions, buffer.String())
	}
	return functions, nil
}

func (g *defaultGenerator) getInterfaceFuncs(service parser.Service) ([]string, error) {
	functions := make([]string, 0)

	for _, rpc := range service.RPC {
		text, err := util.LoadTemplate(category, callInterfaceFunctionTemplateFile, callInterfaceFunctionTemplate)
		if err != nil {
			return nil, err
		}

		comment := parser.GetComment(rpc.Doc())
		buffer, err := util.With("interfaceFn").Parse(text).Execute(
			map[string]interface{}{
				"hasComment": len(comment) > 0,
				"comment":    comment,
				"method":     parser.CamelCase(rpc.Name),
				"pbRequest":  parser.CamelCase(rpc.RequestType),
				"pbResponse": parser.CamelCase(rpc.ReturnsType),
			})
		if err != nil {
			return nil, err
		}

		functions = append(functions, buffer.String())
	}

	return functions, nil
}
