package gen

import (
	"github.com/gotid/god/tools/god/model/sql/template"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
)

func genImports(table Table, withCache, timeImport bool) (string, error) {
	file, builtin := importsWithNoCacheTemplateFile, template.ImportsNoCache
	if withCache {
		file = importsTemplateFile
		builtin = template.Imports
	}

	text, err := pathx.LoadTemplate(category, file, builtin)
	if err != nil {
		return "", err
	}

	buffer, err := util.With("import").Parse(text).Execute(map[string]any{
		"time":       timeImport,
		"containsPQ": table.ContainsPQ,
		"data":       table,
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
