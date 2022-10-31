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

func genDelete(table Table, withCache, postgreSql bool) (string, string, error) {
	keySet := collection.NewSet()
	keyVarSet := collection.NewSet()
	keySet.AddStr(table.PrimaryCacheKey.KeyExpression)
	keyVarSet.AddStr(table.PrimaryCacheKey.KeyLeft)
	for _, key := range table.UniqueCacheKey {
		keySet.AddStr(key.DataKeyExpression)
		keyVarSet.AddStr(key.KeyLeft)
	}
	keys := keySet.KeysStr()
	sort.Strings(keys)
	keyVars := keyVarSet.KeysStr()
	sort.Strings(keyVars)

	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, deleteTemplateFile, template.Delete)
	if err != nil {
		return "", "", err
	}

	output, err := util.With("delete").Parse(text).Execute(map[string]interface{}{
		"upperStartCamelObject":     camel,
		"withCache":                 withCache,
		"containsIndexCache":        table.ContainsUniqueCacheKey,
		"lowerStartCamelPrimaryKey": util.EscapeGolangKeyword(stringx.From(table.PrimaryKey.Name.ToCamel()).UnTitle()),
		"dataType":                  table.PrimaryKey.DataType,
		"keys":                      strings.Join(keys, "\n"),
		"originalPrimaryKey":        wrapWithRawString(table.PrimaryKey.Name.Source(), postgreSql),
		"keyValues":                 strings.Join(keyVars, ", "),
		"postgreSql":                postgreSql,
		"data":                      table,
	})
	if err != nil {
		return "", "", err
	}

	// 接口方法
	text, err = pathx.LoadTemplate(category, deleteMethodTemplateFile, template.DeleteMethod)
	if err != nil {
		return "", "", err
	}

	deleteMethodOut, err := util.With("deleteMethod").
		Parse(text).
		Execute(map[string]interface{}{
			"lowerStartCamelPrimaryKey": util.EscapeGolangKeyword(stringx.From(table.PrimaryKey.Name.ToCamel()).UnTitle()),
			"dataType":                  table.PrimaryKey.DataType,
			"data":                      table,
		})
	if err != nil {
		return "", "", err
	}

	return output.String(), deleteMethodOut.String(), nil
}
