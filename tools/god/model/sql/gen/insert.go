package gen

import (
	"fmt"
	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/tools/god/model/sql/template"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
	"sort"
	"strings"
)

func genInsert(table Table, withCache, postgreSql bool) (string, string, error) {
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

	expressions := make([]string, 0)
	expressionValues := make([]string, 0)
	var count int
	for _, field := range table.Fields {
		camel := util.SafeString(field.Name.ToCamel())
		if table.isIgnoreColumns(field.Name.Source()) {
			continue
		}

		if field.Name.Source() == table.PrimaryKey.Name.Source() {
			if table.PrimaryKey.AutoIncrement {
				continue
			}
		}

		count += 1
		if postgreSql {
			expressions = append(expressions, fmt.Sprintf("$%d", count))
		} else {
			expressions = append(expressions, "?")
		}
		expressionValues = append(expressionValues, "data."+camel)
	}

	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, insertTemplateFile, template.Insert)
	if err != nil {
		return "", "", err
	}

	output, err := util.With("insert").Parse(text).Execute(map[string]interface{}{
		"withCache":             withCache,
		"upperStartCamelObject": camel,
		"lowerStartCamelObject": stringx.From(camel).UnTitle(),
		"expression":            strings.Join(expressions, ", "),
		"expressionValues":      strings.Join(expressionValues, ", "),
		"keys":                  strings.Join(keys, "\n"),
		"keyValues":             strings.Join(keyVars, ", "),
		"data":                  table,
	})
	if err != nil {
		return "", "", err
	}

	// 接口方法
	text, err = pathx.LoadTemplate(category, insertTemplateMethodFile, template.InsertMethod)
	if err != nil {
		return "", "", err
	}

	insertMethodOutput, err := util.With("insertMethod").Parse(text).Execute(map[string]interface{}{
		"upperStartCamelObject": camel,
		"data":                  table,
	})
	if err != nil {
		return "", "", err
	}

	return output.String(), insertMethodOutput.String(), nil
}
