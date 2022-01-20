package generator

import (
	"fmt"
	"path/filepath"

	conf "github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/rpc/parser"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/format"
)

const svcTemplate = `package svc

import {{.imports}}

type ServiceContext struct {
	c config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		c:c,
	}
}
`

func (g *defaultGenerator) GenSvc(ctx DirContext, _ parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetSvc()
	svcFilename, err := format.FileNamingFormat(cfg.NamingFormat, "service_context")
	if err != nil {
		return err
	}

	fileName := filepath.Join(dir.Filename, svcFilename+".go")
	text, err := util.LoadTemplate(category, svcTemplateFile, svcTemplate)
	if err != nil {
		return err
	}

	return util.With("svc").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"imports": fmt.Sprintf(`"%v"`, ctx.GetConfig().Package),
	}, fileName, false)
}
