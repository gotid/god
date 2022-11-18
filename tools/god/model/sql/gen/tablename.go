package gen

import (
	"github.com/gotid/god/tools/god/model/sql/template"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
)

func genTableName(table Table) (string, error) {
	text, err := pathx.LoadTemplate(category, tableNameTemplateFile, template.TableName)
	if err != nil {
		return "", err
	}

	output, err := util.With("tableName").Parse(text).Execute(map[string]any{
		"tableName":             table.Name.Source(),
		"upperStartCamelObject": table.Name.ToCamel(),
	})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
