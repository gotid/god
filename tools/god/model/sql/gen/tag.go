package gen

import (
	"github.com/gotid/god/tools/god/model/sql/template"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
)

func genTag(table Table, fieldName string) (string, error) {
	if fieldName == "" {
		return fieldName, nil
	}

	text, err := pathx.LoadTemplate(category, tagTemplateFile, template.Tag)
	if err != nil {
		return "", err
	}

	output, err := util.With("tag").Parse(text).Execute(map[string]interface{}{
		"field": fieldName,
		"data":  table,
	})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
