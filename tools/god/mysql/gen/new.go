package gen

import (
	"github.com/gotid/god/tools/god/mysql/tpl"
	"github.com/gotid/god/tools/god/util"
)

func genNew(table Table, withCache bool) (string, error) {
	output, err := util.With("new").Parse(tpl.New).Execute(map[string]interface{}{
		"withCache":   withCache,
		"originTable": table.Name.Source(),
		"table":       table.Name.ToCamel(),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
