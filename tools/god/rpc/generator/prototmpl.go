package generator

import (
	_ "embed"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/console"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
	"path/filepath"
	"strings"
)

//go:embed rpc.tpl
var rpcTemplateText string

// ProtoTmpl 输出一个 proto 示例文件至 out 路径。
func ProtoTmpl(out string) error {
	protoFilename := filepath.Base(out)
	serviceName := stringx.From(strings.TrimSuffix(protoFilename, filepath.Ext(protoFilename)))
	text, err := pathx.LoadTemplate(category, rpcTemplateFile, rpcTemplateText)
	if err != nil {
		return err
	}

	dir := filepath.Dir(out)
	err = pathx.MkdirIfNotExist(dir)
	if err != nil {
		return err
	}

	err = util.With("t").Parse(text).SaveTo(map[string]string{
		"package":     serviceName.UnTitle(),
		"serviceName": serviceName.Title(),
	}, out, false)

	console.NewColorConsole().MarkDone()

	return err
}
