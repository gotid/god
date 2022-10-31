package gen

import (
	"fmt"
	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/tools/god/model/sql/template"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
	"strings"
)

func genVars(table Table, withCache, postgreSql bool) (string, error) {
	keys := make([]string, 0)
	keys = append(keys, table.PrimaryCacheKey.VarExpression)
	for _, v := range table.UniqueCacheKey {
		keys = append(keys, v.VarExpression)
	}

	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, varTemplateFile, template.Vars)
	if err != nil {
		return "", err
	}

	output, err := util.With("vars").Parse(text).GoFmt(true).Execute(map[string]interface{}{
		"lowerStartCamelObject": stringx.From(camel).UnTitle(),
		"upperStartCamelObject": camel,
		"cacheKeys":             strings.Join(keys, "\n"),
		"autoIncrement":         table.PrimaryKey.AutoIncrement,
		"originalPrimaryKey":    wrapWithRawString(table.PrimaryKey.Name.Source(), postgreSql),
		"withCache":             withCache,
		"postgreSql":            postgreSql,
		"data":                  table,
		"ignoreColumns": func() string {
			var set = collection.NewSet()
			for _, c := range table.ignoreColumns {
				if postgreSql {
					set.AddStr(fmt.Sprintf(`"%s"`, c))
				} else {
					set.AddStr(fmt.Sprintf("\"`%s`\"", c))
				}
			}
			return strings.Join(set.KeysStr(), ", ")
		}(),
	})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
