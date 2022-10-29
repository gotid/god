package generator

import (
	_ "embed"
	"fmt"
	"github.com/gotid/god/lib/collection"
	conf "github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/rpc/parser"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/format"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
	"path/filepath"
	"strings"
)

const logicFunctionTemplate = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} ({{if .hasReq}}in {{.request}}{{if .stream}},stream {{.streamBody}}{{end}}{{else}}stream {{.streamBody}}{{end}}) ({{if .hasReply}}{{.response}},{{end}} error) {
	//TODO 此处添加你的逻辑并删除该行

	return {{if .hasReply}}&{{.responseType}}{},{{end}} nil
}
`

//go:embed logic.tpl
var logicTemplate string

// GenLogic 生成 rpc 服务的逻辑文件，对应 proto 文件中的 rpc 定义。
func (g *Generator) GenLogic(ctx DirContext, proto parser.Proto, cfg *conf.Config, c *RpcContext) error {
	if !c.Multiple {
		return g.genLogicInCompatibility(ctx, proto, cfg)
	}

	return g.genLogicGroup(ctx, proto, cfg)
}

func (g *Generator) genLogicInCompatibility(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetLogic()
	service := proto.Service[0].Service.Name
	for _, rpc := range proto.Service[0].RPC {
		logicName := fmt.Sprintf("%sLogic", stringx.From(rpc.Name).ToCamel())
		logicFilename, err := format.FileNamingFormat(cfg.NamingFormat, rpc.Name+"_logic")
		if err != nil {
			return err
		}

		filename := filepath.Join(dir.Filename, logicFilename+".go")
		functions, err := g.genLogicFunction(service, proto.PbPackage, logicName, rpc)
		if err != nil {
			return err
		}

		imports := collection.NewSet()
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetSvc().Package))
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetPb().Package))
		text, err := pathx.LoadTemplate(category, logicTemplateFile, logicTemplate)
		if err != nil {
			return err
		}
		if err = util.With("logic").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
			"logicName":   fmt.Sprintf("%sLogic", stringx.From(rpc.Name).ToCamel()),
			"functions":   functions,
			"packageName": "logic",
			"imports":     strings.Join(imports.KeysStr(), pathx.NL),
		}, filename, false); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) genLogicGroup(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetLogic()
	for _, item := range proto.Service {
		serviceName := item.Name
		for _, rpc := range item.RPC {
			var (
				err           error
				filename      string
				logicName     string
				logicFilename string
				packageName   string
			)

			logicName = fmt.Sprintf("%sLogic", stringx.From(rpc.Name).ToCamel())
			childPkg, err := dir.GetChildPackage(serviceName)
			if err != nil {
				return err
			}

			serviceDir := filepath.Base(childPkg)
			nameJoin := fmt.Sprintf("%s_logic", serviceName)
			packageName = strings.ToLower(stringx.From(nameJoin).ToCamel())
			logicFilename, err = format.FileNamingFormat(cfg.NamingFormat, rpc.Name+"_logic")
			if err != nil {
				return err
			}

			filename = filepath.Join(dir.Filename, serviceDir, logicFilename+".go")
			functions, err := g.genLogicFunction(serviceName, proto.PbPackage, logicName, rpc)
			if err != nil {
				return err
			}

			imports := collection.NewSet()
			imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetSvc().Package))
			imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetPb().Package))
			text, err := pathx.LoadTemplate(category, logicTemplateFile, logicTemplate)
			if err != nil {
				return err
			}

			if err = util.With("logic").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
				"logicName":   logicName,
				"functions":   functions,
				"packageName": packageName,
				"imports":     strings.Join(imports.KeysStr(), pathx.NL),
			}, filename, false); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) genLogicFunction(serviceName, goPackage, logicName string, rpc *parser.RPC) (string, error) {
	functions := make([]string, 0)
	text, err := pathx.LoadTemplate(category, logicFuncTemplateFile, logicFunctionTemplate)
	if err != nil {
		return "", err
	}

	comment := parser.GetComment(rpc.Doc())
	streamServer := fmt.Sprintf("%s.%s_%s%s", goPackage, parser.CamelCase(serviceName),
		parser.CamelCase(rpc.Name), "Server")
	buffer, err := util.With("fun").Parse(text).Execute(map[string]interface{}{
		"logicName":    logicName,
		"method":       parser.CamelCase(rpc.Name),
		"hasReq":       !rpc.StreamsRequest,
		"request":      fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.RequestType)),
		"hasReply":     !rpc.StreamsRequest && !rpc.StreamsReturns,
		"response":     fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"responseType": fmt.Sprintf("%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"stream":       rpc.StreamsRequest || rpc.StreamsReturns,
		"streamBody":   streamServer,
		"hasComment":   len(comment) > 0,
		"comment":      comment,
	})
	if err != nil {
		return "", err
	}

	functions = append(functions, buffer.String())
	return strings.Join(functions, pathx.NL), nil
}
