package gen

import (
	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/tools/god/model/sql/template"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
	"sort"
	"strings"
)

func genUpdate(table Table, withCache, postgreSql bool) (string, string, error) {
	expressionValues := make([]string, 0)
	pkg := "data."
	if table.ContainsUniqueCacheKey {
		pkg = "newData."
	}
	for _, field := range table.Fields {
		camel := util.SafeString(field.Name.ToCamel())
		if table.isIgnoreColumns(field.Name.Source()) {
			continue
		}

		if field.Name.Source() == table.PrimaryKey.Name.Source() {
			continue
		}

		expressionValues = append(expressionValues, pkg+camel)
	}

	keySet := collection.NewSet()
	keyVarSet := collection.NewSet()
	keySet.AddStr(table.PrimaryCacheKey.DataKeyExpression)
	keyVarSet.AddStr(table.PrimaryCacheKey.KeyLeft)
	for _, key := range table.UniqueCacheKey {
		keySet.AddStr(key.DataKeyExpression)
		keyVarSet.AddStr(key.KeyLeft)
	}
	keys := keySet.KeysStr()
	sort.Strings(keys)
	keyVars := keyVarSet.KeysStr()
	sort.Strings(keyVars)

	if postgreSql {
		expressionValues = append([]string{pkg + table.PrimaryKey.Name.ToCamel()}, expressionValues...)
	} else {
		expressionValues = append(expressionValues, pkg+table.PrimaryKey.Name.ToCamel())
	}
	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, updateTemplateFile, template.Update)
	if err != nil {
		return "", "", err
	}

	output, err := util.With("update").Parse(text).Execute(map[string]any{
		"withCache":             withCache,
		"containsIndexCache":    table.ContainsUniqueCacheKey,
		"upperStartCamelObject": camel,
		"keys":                  strings.Join(keys, "\n"),
		"keyValues":             strings.Join(keyVars, ", "),
		"primaryCacheKey":       table.PrimaryCacheKey.DataKeyExpression,
		"primaryKeyVariable":    table.PrimaryCacheKey.KeyLeft,
		"lowerStartCamelObject": stringx.From(camel).UnTitle(),
		"upperStartCamelPrimaryKey": util.EscapeGolangKeyword(
			stringx.From(table.PrimaryKey.Name.ToCamel()).Title(),
		),
		"originalPrimaryKey": wrapWithRawString(
			table.PrimaryKey.Name.Source(), postgreSql,
		),
		"expressionValues": strings.Join(
			expressionValues, ", ",
		),
		"postgreSql": postgreSql,
		"data":       table,
	})
	if err != nil {
		return "", "", err
	}

	// 更新接口方法
	text, err = pathx.LoadTemplate(category, updateMethodTemplateFile, template.UpdateMethod)
	if err != nil {
		return "", "", err
	}

	updateMethodOutput, err := util.With("updateMethod").Parse(text).Execute(map[string]any{
		"upperStartCamelObject": camel,
		"data":                  table,
	})
	if err != nil {
		return "", "", err
	}

	return output.String(), updateMethodOutput.String(), nil
}
