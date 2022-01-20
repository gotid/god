package gen

import (
	"github.com/gotid/god/lib/stringx"
	"github.com/gotid/god/tools/god/mysql/tpl"
	"github.com/gotid/god/tools/god/util"
)

func genFindOne(table Table, withCache bool) (string, error) {
	upperTable := table.Name.ToCamel()
	output, err := util.With("findOne").Parse(tpl.FindOne).Execute(map[string]interface{}{
		"withCache":          withCache,
		"upperTable":         upperTable,
		"lowerTable":         stringx.From(upperTable).UnTitle(),
		"originalPrimaryKey": table.PrimaryKey.Name.Source(),
		"primaryKey":         stringx.From(table.PrimaryKey.Name.ToCamel()).UnTitle(),
		"dataType":           table.PrimaryKey.DataType,
		"cacheKeyName":       table.CacheKeys[table.PrimaryKey.Name.Source()].KeyName,
		"cacheKeyExpression": table.CacheKeys[table.PrimaryKey.Name.Source()].KeyExpression,
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
