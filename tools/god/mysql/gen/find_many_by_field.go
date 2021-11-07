package gen

import (
	"strings"

	"git.zc0901.com/go/god/tools/god/mysql/tpl"
	"git.zc0901.com/go/god/tools/god/util"
)

func genFindManyByFields(table Table) (string, error) {
	t := util.With("findManyByFields").Parse(tpl.FindManyByFields)
	var list []string
	upperTable := table.Name.ToCamel()
	for _, field := range table.Fields {
		if field.IsPrimaryKey || !field.IsUniqueKey {
			continue
		}
		upperField := field.Name.ToCamel()
		output, err := t.Execute(map[string]interface{}{
			"upperTable": upperTable,
			"upperField": upperField,
			"dataType":   field.DataType,
		})
		if err != nil {
			return "", err
		}
		list = append(list, output.String())
	}
	return strings.Join(list, "\n"), nil
}
