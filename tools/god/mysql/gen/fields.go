package gen

import (
	"strings"

	"github.com/gotid/god/tools/god/mysql/parser"
	"github.com/gotid/god/tools/god/mysql/tpl"
	"github.com/gotid/god/tools/god/util"
)

func genFields(fields []parser.Field) (string, error) {
	var list []string
	for _, field := range fields {
		result, err := genField(field)
		if err != nil {
			return "", err
		}
		list = append(list, result)
	}
	return strings.Join(list, "\n"), nil
}

func genField(field parser.Field) (string, error) {
	tag, err := genTag(field.Name.Source())
	if err != nil {
		return "", err
	}
	output, err := util.With("types").
		Parse(tpl.Field).
		Execute(map[string]interface{}{
			"name":       field.Name.ToCamel(),
			"type":       field.DataType,
			"tag":        tag,
			"hasComment": field.Comment != "",
			"comment":    field.Comment,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
