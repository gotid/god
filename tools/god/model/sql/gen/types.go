package gen

import (
	"github.com/gotid/god/tools/god/model/sql/template"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
)

func genTypes(table Table, methods string, withCache bool) (string, error) {
	fields := table.Fields
	fieldsString, err := genFields(table, fields)
	if err != nil {
		return "", err
	}

	text, err := pathx.LoadTemplate(category, typesTemplateFile, template.Types)
	if err != nil {
		return "", err
	}

	output, err := util.With("types").Parse(text).Execute(map[string]any{
		"withCache":             withCache,
		"method":                methods,
		"upperStartCamelObject": table.Name.ToCamel(),
		"lowerStartCamelObject": stringx.From(table.Name.ToCamel()).UnTitle(),
		"fields":                fieldsString,
		"data":                  table,
	})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
