package gen

import (
	"strings"

	"github.com/gotid/god/lib/stringx"
	"github.com/gotid/god/tools/god/mysql/tpl"
	"github.com/gotid/god/tools/god/util"
)

func genUpdate(table Table, withCache bool) (string, error) {
	values := make([]string, 0)
	for _, field := range table.Fields {
		upperField := field.Name.ToCamel()
		if field.IsPrimaryKey || upperField == "CreatedAt" || upperField == "UpdatedAt" || upperField == "CreateTime" || upperField == "UpdateTime" {
			continue
		}
		values = append(values, "data."+upperField)
	}

	values = append(values, "data."+table.PrimaryKey.Name.ToCamel())
	upperTable := table.Name.ToCamel()
	output, err := util.With("update").
		Parse(tpl.Update).
		Execute(map[string]interface{}{
			"withCache":          withCache,
			"upperTable":         upperTable,
			"primaryCacheKey":    table.CacheKeys[table.PrimaryKey.Name.Source()].DataKeyExpression,
			"primaryKeyName":     table.CacheKeys[table.PrimaryKey.Name.Source()].KeyName,
			"lowerTable":         stringx.From(upperTable).UnTitle(),
			"originalPrimaryKey": table.PrimaryKey.Name.Source(),
			"values":             strings.Join(values, ", "),
		})
	if err != nil {
		return "", nil
	}
	return output.String(), nil
}

func genUpdatePartial(table Table, withCache bool) (string, error) {
	values := make([]string, 0)
	for _, field := range table.Fields {
		upperField := field.Name.ToCamel()
		if field.IsPrimaryKey || upperField == "CreatedAt" || upperField == "UpdatedAt" || upperField == "CreateTime" || upperField == "UpdateTime" {
			continue
		}
		values = append(values, "data."+upperField)
	}

	values = append(values, "data."+table.PrimaryKey.Name.ToCamel())
	upperTable := table.Name.ToCamel()
	output, err := util.With("update").
		Parse(tpl.UpdatePartial).
		Execute(map[string]interface{}{
			"withCache":          withCache,
			"upperTable":         upperTable,
			"primaryCacheKey":    strings.ReplaceAll(table.CacheKeys[table.PrimaryKey.Name.Source()].DataKeyExpression, "data.Id", "updateArgs.Id"),
			"primaryKeyName":     table.CacheKeys[table.PrimaryKey.Name.Source()].KeyName,
			"lowerTable":         stringx.From(upperTable).UnTitle(),
			"originalPrimaryKey": table.PrimaryKey.Name.Source(),
		})
	if err != nil {
		return "", nil
	}
	return output.String(), nil
}

func genTxUpdate(table Table, withCache bool) (string, error) {
	values := make([]string, 0)
	for _, field := range table.Fields {
		upperField := field.Name.ToCamel()
		if field.IsPrimaryKey || upperField == "CreatedAt" || upperField == "UpdatedAt" || upperField == "CreateTime" || upperField == "UpdateTime" {
			continue
		}
		values = append(values, "data."+upperField)
	}

	values = append(values, "data."+table.PrimaryKey.Name.ToCamel())
	upperTable := table.Name.ToCamel()
	output, err := util.With("update").
		Parse(tpl.TxUpdate).
		Execute(map[string]interface{}{
			"withCache":          withCache,
			"upperTable":         upperTable,
			"primaryCacheKey":    table.CacheKeys[table.PrimaryKey.Name.Source()].DataKeyExpression,
			"primaryKeyName":     table.CacheKeys[table.PrimaryKey.Name.Source()].KeyName,
			"lowerTable":         stringx.From(upperTable).UnTitle(),
			"originalPrimaryKey": table.PrimaryKey.Name.Source(),
			"values":             strings.Join(values, ", "),
		})
	if err != nil {
		return "", nil
	}
	return output.String(), nil
}

func genTxUpdatePartial(table Table, withCache bool) (string, error) {
	values := make([]string, 0)
	for _, field := range table.Fields {
		upperField := field.Name.ToCamel()
		if field.IsPrimaryKey || upperField == "CreatedAt" || upperField == "UpdatedAt" || upperField == "CreateTime" || upperField == "UpdateTime" {
			continue
		}
		values = append(values, "data."+upperField)
	}

	values = append(values, "data."+table.PrimaryKey.Name.ToCamel())
	upperTable := table.Name.ToCamel()
	output, err := util.With("update").
		Parse(tpl.TxUpdatePartial).
		Execute(map[string]interface{}{
			"withCache":          withCache,
			"upperTable":         upperTable,
			"primaryCacheKey":    strings.ReplaceAll(table.CacheKeys[table.PrimaryKey.Name.Source()].DataKeyExpression, "data.Id", "updateArgs.Id"),
			"primaryKeyName":     table.CacheKeys[table.PrimaryKey.Name.Source()].KeyName,
			"lowerTable":         stringx.From(upperTable).UnTitle(),
			"originalPrimaryKey": table.PrimaryKey.Name.Source(),
		})
	if err != nil {
		return "", nil
	}
	return output.String(), nil
}
