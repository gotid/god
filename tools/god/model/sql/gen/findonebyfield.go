package gen

import (
	"fmt"
	"github.com/gotid/god/tools/god/model/sql/template"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
	"strings"
)

type findOneCode struct {
	findOneMethod          string
	findOneInterfaceMethod string
	cacheExtra             string
}

func genFindOneByField(table Table, withCache, postgreSql bool) (*findOneCode, error) {
	text, err := pathx.LoadTemplate(category, findOneByFieldTemplateFile, template.FindOneByField)
	if err != nil {
		return nil, err
	}

	t := util.With("findOneByField").Parse(text)
	var list []string
	camel := table.Name.ToCamel()
	for _, key := range table.UniqueCacheKey {
		in, paramJoinString, originalFieldString := convertJoin(key, postgreSql)

		output, err := t.Execute(map[string]any{
			"upperStartCamelObject":     camel,
			"upperField":                key.FieldNameJoin.ToCamel().With("").Source(),
			"in":                        in,
			"withCache":                 withCache,
			"cacheKey":                  key.KeyExpression,
			"cacheKeyVariable":          key.KeyLeft,
			"lowerStartCamelObject":     stringx.From(camel).UnTitle(),
			"lowerStartCamelField":      paramJoinString,
			"upperStartCamelPrimaryKey": table.PrimaryKey.Name.ToCamel(),
			"originalField":             originalFieldString,
			"postgreSql":                postgreSql,
			"data":                      table,
		})
		if err != nil {
			return nil, err
		}

		list = append(list, output.String())
	}

	text, err = pathx.LoadTemplate(category, findOneByFieldMethodTemplateFile, template.FindOneByFieldMethod)
	if err != nil {
		return nil, err
	}

	t = util.With("findOneByFieldMethod").Parse(text)
	var listMethod []string
	for _, key := range table.UniqueCacheKey {
		var inJoin, paramJoin Join
		for _, field := range key.Fields {
			param := util.EscapeGolangKeyword(stringx.From(field.Name.ToCamel()).UnTitle())
			inJoin = append(inJoin, fmt.Sprintf("%s %s", param, field.DataType))
			paramJoin = append(paramJoin, param)
		}

		var in string
		if len(inJoin) > 0 {
			in = inJoin.With(", ").Source()
		}
		output, err := t.Execute(map[string]any{
			"upperStartCamelObject": camel,
			"upperField":            key.FieldNameJoin.ToCamel().With("").Source(),
			"in":                    in,
			"data":                  table,
		})
		if err != nil {
			return nil, err
		}

		listMethod = append(listMethod, output.String())
	}

	if withCache {
		text, err := pathx.LoadTemplate(category, findOneByFieldExtraMethodTemplateFile, template.FindOneByFieldExtraMethod)
		if err != nil {
			return nil, err
		}

		out, err := util.With("findOneByFieldExtraMethod").Parse(text).Execute(map[string]any{
			"upperStartCamelObject": camel,
			"primaryKeyLeft":        table.PrimaryCacheKey.VarLeft,
			"lowerStartCamelObject": stringx.From(camel).UnTitle(),
			"originalPrimaryField":  wrapWithRawString(table.PrimaryKey.Name.Source(), postgreSql),
			"postgreSql":            postgreSql,
			"data":                  table,
		})
		if err != nil {
			return nil, err
		}

		return &findOneCode{
			findOneMethod:          strings.Join(list, pathx.NL),
			findOneInterfaceMethod: strings.Join(listMethod, pathx.NL),
			cacheExtra:             out.String(),
		}, nil
	}

	return &findOneCode{
		findOneMethod:          strings.Join(list, pathx.NL),
		findOneInterfaceMethod: strings.Join(listMethod, pathx.NL),
	}, nil
}

func convertJoin(key Key, postgreSql bool) (in, paramJoinString, originalFieldString string) {
	var inJoin, paramJoin, argJoin Join
	for i, field := range key.Fields {
		param := util.EscapeGolangKeyword(stringx.From(field.Name.ToCamel()).UnTitle())
		inJoin = append(inJoin, fmt.Sprintf("%s %s", param, field.DataType))
		paramJoin = append(paramJoin, param)
		if postgreSql {
			argJoin = append(argJoin, fmt.Sprintf("%s = $%d", wrapWithRawString(field.Name.Source(), postgreSql), i+1))
		} else {
			argJoin = append(argJoin, fmt.Sprintf("%s = ?", wrapWithRawString(field.Name.Source(), postgreSql)))
		}
	}

	if len(inJoin) > 0 {
		in = inJoin.With(", ").Source()
	}
	if len(paramJoin) > 0 {
		paramJoinString = paramJoin.With(",").Source()
	}
	if len(argJoin) > 0 {
		originalFieldString = argJoin.With(" and ").Source()
	}

	return in, paramJoinString, originalFieldString
}
