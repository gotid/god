package gen

import (
	"git.zc0901.com/go/god/lib/stringx"
	"git.zc0901.com/go/god/tools/god/mysql/tpl"
	"git.zc0901.com/go/god/tools/god/util"
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
