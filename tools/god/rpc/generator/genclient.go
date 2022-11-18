package generator

import (
	_ "embed"
	"fmt"
	"github.com/emicklei/proto"
	"github.com/gotid/god/lib/collection"
	conf "github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/rpc/parser"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/format"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
	"path/filepath"
	"sort"
	"strings"
)

const (
	clientInterfaceFunctionTemplate = `{{if .hasComment}}{{.comment}}
{{end}}{{.method}}(ctx context.Context{{if .hasReq}}, in *{{.pbRequest}}{{end}}, opts ...grpc.CallOption) ({{if .notStream}}*{{.pbResponse}}, {{else}}{{.streamBody}},{{end}} error)`

	clientFunctionTemplate = `
{{if .hasComment}}{{.comment}}{{end}}
func (m *default{{.serviceName}}) {{.method}}(ctx context.Context{{if .hasReq}}, in *{{.pbRequest}}{{end}}, opts ...grpc.CallOption) ({{if .notStream}}*{{.pbResponse}}, {{else}}{{.streamBody}},{{end}} error) {
	client := {{if .isCallPkgSameToGrpcPkg}}{{else}}{{.package}}.{{end}}New{{.rpcServiceName}}Client(m.cli.Conn())
	return client.{{.method}}(ctx{{if .hasReq}}, in{{end}}, opts...)
}
`
)

//go:embed client.tpl
var clientTemplateText string

func (g *Generator) GenClient(ctx DirContext, proto parser.Proto, cfg *conf.Config, c *RpcContext) error {
	if !c.Multiple {
		return g.genClientInCompatibility(ctx, proto, cfg)
	}

	return g.genClientGroup(ctx, proto, cfg)
}

func (g *Generator) genClientGroup(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetClient()
	head := util.GetHead(proto.Name)
	for _, service := range proto.Service {
		childPkg, err := dir.GetChildPackage(service.Name)
		if err != nil {
			return err
		}

		clientFilename, err := format.FileNamingFormat(cfg.NamingFormat, service.Name)
		if err != nil {
			return err
		}

		childDir := filepath.Base(childPkg)
		filename := filepath.Join(dir.Filename, childDir, fmt.Sprintf("%s.go", clientFilename))
		isCallPkgSameToPbPkg := childDir == ctx.GetProtoGo().Filename
		isCallPkgSameToGrpcPkg := childDir == ctx.GetProtoGo().Filename

		functions, err := g.genCallFunctions(proto.PbPackage, service, isCallPkgSameToGrpcPkg)
		if err != nil {
			return err
		}

		iFunctions, err := g.getInterfaceFunctions(proto.PbPackage, service, isCallPkgSameToGrpcPkg)
		if err != nil {
			return err
		}

		text, err := pathx.LoadTemplate(category, clientTemplateFile, clientTemplateText)
		if err != nil {
			return err
		}

		alias := collection.NewSet()
		if !isCallPkgSameToPbPkg {
			for _, item := range proto.Message {
				msgName := getMessageName(*item.Message)
				alias.AddStr(fmt.Sprintf("%s = %s", parser.CamelCase(msgName),
					fmt.Sprintf("%s.%s", proto.PbPackage, parser.CamelCase(msgName))))
			}
		}

		pbPackage := fmt.Sprintf(`"%s"`, ctx.GetPb().Package)
		protoGoPackage := fmt.Sprintf(`"%s"`, ctx.GetProtoGo().Package)
		if isCallPkgSameToGrpcPkg {
			pbPackage = ""
			protoGoPackage = ""
		}

		aliasKeys := alias.KeysStr()
		sort.Strings(aliasKeys)
		if err = util.With("shared").GoFmt(true).Parse(text).SaveTo(map[string]any{
			"name":           clientFilename,
			"alias":          strings.Join(aliasKeys, pathx.NL),
			"head":           head,
			"filePackage":    dir.Base,
			"pbPackage":      pbPackage,
			"protoGoPackage": protoGoPackage,
			"serviceName":    stringx.From(service.Name).ToCamel(),
			"functions":      strings.Join(functions, pathx.NL),
			"interface":      strings.Join(iFunctions, pathx.NL),
		}, filename, true); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) genClientInCompatibility(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetClient()
	service := proto.Service[0]
	head := util.GetHead(proto.Name)
	isCallPkgSameToPbPkg := ctx.GetClient().Filename == ctx.GetPb().Filename
	isCallPkgSameToGrpcPkg := ctx.GetClient().Filename == ctx.GetProtoGo().Filename

	clientFilename, err := format.FileNamingFormat(cfg.NamingFormat, service.Name)
	if err != nil {
		return err
	}

	filename := filepath.Join(dir.Filename, fmt.Sprintf("%s.go", clientFilename))
	functions, err := g.genCallFunctions(proto.PbPackage, service, isCallPkgSameToGrpcPkg)
	if err != nil {
		return err
	}

	iFunctions, err := g.getInterfaceFunctions(proto.PbPackage, service, isCallPkgSameToGrpcPkg)
	if err != nil {
		return err
	}

	text, err := pathx.LoadTemplate(category, clientTemplateFile, clientTemplateText)
	if err != nil {
		return err
	}

	alias := collection.NewSet()
	if !isCallPkgSameToPbPkg {
		for _, item := range proto.Message {
			msgName := getMessageName(*item.Message)
			alias.AddStr(fmt.Sprintf("%s = %s", parser.CamelCase(msgName),
				fmt.Sprintf("%s.%s", proto.PbPackage, parser.CamelCase(msgName))))
		}
	}

	pbPackage := fmt.Sprintf(`"%s"`, ctx.GetPb().Package)
	protoGoPackage := fmt.Sprintf(`"%s"`, ctx.GetProtoGo().Package)
	if isCallPkgSameToGrpcPkg {
		pbPackage = ""
		protoGoPackage = ""
	}
	aliasKeys := alias.KeysStr()
	sort.Strings(aliasKeys)
	return util.With("shared").GoFmt(true).Parse(text).SaveTo(map[string]any{
		"name":           clientFilename,
		"alias":          strings.Join(aliasKeys, pathx.NL),
		"head":           head,
		"filePackage":    dir.Base,
		"pbPackage":      pbPackage,
		"protoGoPackage": protoGoPackage,
		"serviceName":    stringx.From(service.Name).ToCamel(),
		"functions":      strings.Join(functions, pathx.NL),
		"interface":      strings.Join(iFunctions, pathx.NL),
	}, filename, true)
}

func getMessageName(msg proto.Message) string {
	list := []string{msg.Name}

	for {
		parent := msg.Parent
		if parent == nil {
			break
		}

		parentMsg, ok := parent.(*proto.Message)
		if !ok {
			break
		}

		tmp := []string{parentMsg.Name}
		list = append(tmp, list...)
		msg = *parentMsg
	}

	return strings.Join(list, "_")
}

func (g *Generator) getInterfaceFunctions(pbPackage string, service parser.Service, isCallPkgSameToGrpcPkg bool) ([]string, error) {
	functions := make([]string, 0)

	for _, rpc := range service.RPC {
		text, err := pathx.LoadTemplate(category, clientInterfaceFunctionTemplateFile, clientInterfaceFunctionTemplate)
		if err != nil {
			return nil, err
		}

		comment := parser.GetComment(rpc.Doc())
		streamServer := fmt.Sprintf("%s.%s_%s%s", pbPackage, parser.CamelCase(service.Name),
			parser.CamelCase(rpc.Name), "Client")
		if isCallPkgSameToGrpcPkg {
			streamServer = fmt.Sprintf("%s_%s%s", parser.CamelCase(service.Name),
				parser.CamelCase(rpc.Name), "Client")
		}

		buffer, err := util.With("interfaceFn").Parse(text).Execute(map[string]any{
			"hasComment": len(comment) > 0,
			"comment":    comment,
			"method":     parser.CamelCase(rpc.Name),
			"hasReq":     !rpc.StreamsRequest,
			"pbRequest":  parser.CamelCase(rpc.RequestType),
			"notStream":  !rpc.StreamsRequest && !rpc.StreamsReturns,
			"pbResponse": parser.CamelCase(rpc.ReturnsType),
			"streamBody": streamServer,
		})
		if err != nil {
			return nil, err
		}

		functions = append(functions, buffer.String())
	}

	return functions, nil
}

func (g *Generator) genCallFunctions(pbPackage string, service parser.Service, isCallPkgSameToGrpcPkg bool) ([]string, error) {
	functions := make([]string, 0)

	for _, rpc := range service.RPC {
		text, err := pathx.LoadTemplate(category, clientFunctionTemplateFile, clientFunctionTemplate)
		if err != nil {
			return nil, err
		}

		comment := parser.GetComment(rpc.Doc())
		streamServer := fmt.Sprintf("%s.%s_%s%s", pbPackage, parser.CamelCase(service.Name), parser.CamelCase(rpc.Name), "Client")
		if isCallPkgSameToGrpcPkg {
			streamServer = fmt.Sprintf("%s_%s%s", parser.CamelCase(service.Name), parser.CamelCase(rpc.Name), "Client")
		}
		buffer, err := util.With("sharedFn").Parse(text).Execute(map[string]any{
			"serviceName":            stringx.From(service.Name).ToCamel(),
			"rpcServiceName":         parser.CamelCase(service.Name),
			"method":                 parser.CamelCase(rpc.Name),
			"package":                pbPackage,
			"pbRequest":              parser.CamelCase(rpc.RequestType),
			"pbResponse":             parser.CamelCase(rpc.ReturnsType),
			"hasComment":             len(comment) > 0,
			"comment":                comment,
			"hasReq":                 !rpc.StreamsRequest,
			"notStream":              !rpc.StreamsRequest && !rpc.StreamsReturns,
			"streamBody":             streamServer,
			"isCallPkgSameToGrpcPkg": isCallPkgSameToGrpcPkg,
		})
		if err != nil {
			return nil, err
		}

		functions = append(functions, buffer.String())
	}

	return functions, nil
}
