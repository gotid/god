package generator

import (
	_ "embed"
	conf "github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/rpc/parser"
	"github.com/gotid/god/tools/god/util/format"
	"github.com/gotid/god/tools/god/util/pathx"
	"os"
	"path/filepath"
)

//go:embed config.tpl
var configTemplate string

// GenConfig 生成 rpc 服务的配置结构定义文件。
// 包括 rpc.ServerConfig 默认配置项。
// 你可以通过 config.Config 指定目标文件的命名风格。
// 详见：https://github.com/gotid/god/tree/master/tools/god/config/config.go
func (g *Generator) GenConfig(ctx DirContext, _ parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetConfig()
	configFilename, err := format.FileNamingFormat(cfg.NamingFormat, "config")
	if err != nil {
		return err
	}

	filename := filepath.Join(dir.Filename, configFilename+".go")
	if pathx.FileExists(filename) {
		return nil
	}

	text, err := pathx.LoadTemplate(category, configTemplateFile, configTemplate)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(text), os.ModePerm)
}
