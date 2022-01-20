package gen

import (
	"strings"

	"github.com/gotid/god/lib/container/garray"

	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/lib/stringx"
	"github.com/gotid/god/tools/god/mysql/tpl"
	"github.com/gotid/god/tools/god/util"
)

func genDelete(table Table, withCache bool) (string, error) {
	containsIndexCache := false
	for _, item := range table.Fields {
		if item.IsUniqueKey {
			containsIndexCache = true
			break
		}
	}

	keySet := garray.New()
	keyNamesSet := collection.NewSet()
	if withCache && containsIndexCache {
		keySet.Append("keys := make([]string, len(id)*2)")
	} else {
		keySet.Append("keys := make([]string, len(id))")
	}
	keySet.Append("for i, v := range id {")
	if withCache && containsIndexCache {
		keySet.Append("data := datas[i]")
	}

	for fieldName, key := range table.CacheKeys {
		if fieldName == table.PrimaryKey.Name.Source() {
			keySet.Append(strings.ReplaceAll(
				strings.ReplaceAll(key.KeyExpression, key.KeyName+" :=", "keys[i] ="),
				", id)",
				", v)",
			))
		} else {
			keySet.Append(strings.ReplaceAll(key.DataKeyExpression, key.KeyName+" :=", "keys[i+1] ="))
		}
		keyNamesSet.AddStr(key.KeyName)
	}
	keySet.Append("}")

	upperTable := table.Name.ToCamel()
	output, err := util.With("delete").Parse(tpl.Delete).Execute(map[string]interface{}{
		"upperStartCamelObject":     upperTable,
		"withCache":                 withCache,
		"containsIndexCache":        containsIndexCache,
		"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).UnTitle(),
		"dataType":                  table.PrimaryKey.DataType,
		"keys":                      keySet.Join("\n"),
		"originalPrimaryKey":        table.PrimaryKey.Name.Source(),
		"keyNames":                  strings.Join(keyNamesSet.KeysStr(), ", "),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

func genTxDelete(table Table, withCache bool) (string, error) {
	containsIndexCache := false
	for _, item := range table.Fields {
		if item.IsUniqueKey {
			containsIndexCache = true
			break
		}
	}

	keySet := garray.New()
	keyNamesSet := collection.NewSet()
	if withCache && containsIndexCache {
		keySet.Append("keys := make([]string, len(id)*2)")
	} else {
		keySet.Append("keys := make([]string, len(id))")
	}
	keySet.Append("for i, v := range id {")
	if withCache && containsIndexCache {
		keySet.Append("data := datas[i]")
	}

	for fieldName, key := range table.CacheKeys {
		if fieldName == table.PrimaryKey.Name.Source() {
			keySet.Append(strings.ReplaceAll(
				strings.ReplaceAll(key.KeyExpression, key.KeyName+" :=", "keys[i] ="),
				", id)",
				", v)",
			))
		} else {
			keySet.Append(strings.ReplaceAll(key.DataKeyExpression, key.KeyName+" :=", "keys[i+1] ="))
		}
		keyNamesSet.AddStr(key.KeyName)
	}
	keySet.Append("}")

	upperTable := table.Name.ToCamel()
	output, err := util.With("delete").Parse(tpl.TxDelete).Execute(map[string]interface{}{
		"upperStartCamelObject":     upperTable,
		"withCache":                 withCache,
		"containsIndexCache":        containsIndexCache,
		"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).UnTitle(),
		"dataType":                  table.PrimaryKey.DataType,
		"keys":                      keySet.Join("\n"),
		"originalPrimaryKey":        table.PrimaryKey.Name.Source(),
		"keyNames":                  strings.Join(keyNamesSet.KeysStr(), ", "),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

//func genTxDelete(table Table, withCache bool) (string, error) {
//	keySet := collection.NewSet()
//	keyNamesSet := collection.NewSet()
//	for fieldName, key := range table.CacheKeys {
//		if fieldName == table.PrimaryKey.Name.Source() {
//			keySet.AddStr(key.KeyExpression)
//		} else {
//			keySet.AddStr(key.DataKeyExpression)
//		}
//		keyNamesSet.AddStr(key.KeyName)
//	}
//	containsIndexCache := false
//	for _, item := range table.Fields {
//		if item.IsUniqueKey {
//			containsIndexCache = true
//			break
//		}
//	}
//	upperTable := table.Name.ToCamel()
//	output, err := util.With("delete").Parse(tpl.TxDelete).Execute(map[string]interface{}{
//		"upperStartCamelObject":     upperTable,
//		"withCache":                 withCache,
//		"containsIndexCache":        containsIndexCache,
//		"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).UnTitle(),
//		"dataType":                  table.PrimaryKey.DataType,
//		"keys":                      strings.Join(keySet.KeysStr(), "\n"),
//		"originalPrimaryKey":        table.PrimaryKey.Name.Source(),
//		"keyNames":                  strings.Join(keyNamesSet.KeysStr(), ", "),
//	})
//	if err != nil {
//		return "", err
//	}
//	return output.String(), nil
//}
