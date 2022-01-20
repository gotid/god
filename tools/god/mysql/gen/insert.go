package gen

import (
	"strings"

	"github.com/gotid/god/lib/stringx"
	"github.com/gotid/god/tools/god/mysql/tpl"
	"github.com/gotid/god/tools/god/util"
)

func genInsert(table Table, withCache bool) (string, error) {
	args := make([]string, 0)
	values := make([]string, 0)

	for _, field := range table.Fields {
		camelField := field.Name.ToCamel()
		if camelField == "CreatedAt" || camelField == "UpdatedAt" || camelField == "CreateTime" || camelField == "UpdateTime" {
			continue
		}
		if field.IsPrimaryKey && table.PrimaryKey.AutoIncrement {
			continue
		}

		args = append(args, "?")
		values = append(values, "data."+camelField)
	}
	upperTable := table.Name.ToCamel()
	output, err := util.With("insert").Parse(tpl.Insert).Execute(map[string]interface{}{
		"withCache":  withCache,
		"upperTable": upperTable,
		"lowerTable": stringx.From(upperTable).UnTitle(),
		"args":       strings.Join(args, ", "),
		"values":     strings.Join(values, ", "),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

func genTxInsert(table Table, withCache bool) (string, error) {
	args := make([]string, 0)
	values := make([]string, 0)

	for _, field := range table.Fields {
		camelField := field.Name.ToCamel()
		if camelField == "CreatedAt" || camelField == "UpdatedAt" || camelField == "CreateTime" || camelField == "UpdateTime" {
			continue
		}
		if field.IsPrimaryKey && table.PrimaryKey.AutoIncrement {
			continue
		}

		args = append(args, "?")
		values = append(values, "data."+camelField)
	}
	upperTable := table.Name.ToCamel()
	output, err := util.With("insert").Parse(tpl.TxInsert).Execute(map[string]interface{}{
		"upperTable": upperTable,
		"lowerTable": stringx.From(upperTable).UnTitle(),
		"args":       strings.Join(args, ", "),
		"values":     strings.Join(values, ", "),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
