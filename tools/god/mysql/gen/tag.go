package gen

import (
	"github.com/gotid/god/lib/stringx"
	"github.com/gotid/god/tools/god/mysql/tpl"
	"github.com/gotid/god/tools/god/util"
)

func genTag(fieldName string) (string, error) {
	if fieldName == "" {
		return fieldName, nil
	}

	output, err := util.With("tag").Parse(tpl.Tag).Execute(map[string]interface{}{
		"field":      fieldName,
		"fieldCamel": stringx.From(stringx.From(fieldName).ToCamel()).UnTitle(),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
