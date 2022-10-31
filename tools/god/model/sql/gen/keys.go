package gen

import (
	"fmt"
	"github.com/gotid/god/tools/god/model/sql/parser"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/stringx"
	"sort"
	"strings"
)

// Key 描述缓存键。
type Key struct {
	// VarLeft 形如 cacheUserIdPrefix 的缓存键表达式
	VarLeft string
	// VarRight 形如 "cache:user:id:" 的缓存键表达式
	VarRight string
	// VarExpression 形如 cacheUserIdPrefix = "cache:user:id:" 的缓存键表达式
	VarExpression string
	// KeyLeft 形如 userKey 的键定义表达式
	KeyLeft string
	// KeyRight 形如 fmt.Sprintf("%s%v", cacheUserPrefix, user) 的键定义表达式
	KeyRight string
	// DataKeyRight 形如 fmt.Sprintf("%s%v", cacheUserPrefix, data.User) 的数据键
	DataKeyRight string
	// KeyExpression 形如 userKey := fmt.Sprintf("%s%v", cacheUserPrefix, user) 的键表达式
	KeyExpression string
	// DataKeyExpression 形如 userKey := fmt.Sprintf("%s%v", cacheUserPrefix, data.User) 的键表达式
	DataKeyExpression string
	// FieldNameJoin 描述表的列名切片
	FieldNameJoin Join
	// 描述标的列切片
	Fields []*parser.Field
}

// Join 是 字符串切片的别名。
type Join []string

func genCacheKeys(table parser.Table) (Key, []Key) {
	var primaryKey Key
	var uniqueKey []Key
	primaryKey = genCacheKey(table.Db, table.Name, []*parser.Field{&table.PrimaryKey.Field})
	for _, fields := range table.UniqueIndex {
		uniqueKey = append(uniqueKey, genCacheKey(table.Db, table.Name, fields))
	}
	sort.Slice(uniqueKey, func(i, j int) bool {
		return uniqueKey[i].VarLeft < uniqueKey[j].VarLeft
	})

	return primaryKey, uniqueKey
}

func genCacheKey(db, table stringx.String, fields []*parser.Field) Key {
	var (
		varLeftJoin, varRightJoin, fieldNameJoin Join
		varLeft, varRight, varExpression         string

		keyLeftJoin, keyRightJoin, keyRightArgJoin, dataRightJoin         Join
		keyLeft, keyRight, dataKeyRight, keyExpression, dataKeyExpression string
	)

	dbName, tableName := util.SafeString(db.Source()), util.SafeString(table.Source())
	if len(dbName) > 0 {
		varLeftJoin = append(varLeftJoin, "cache", dbName, tableName)
		varRightJoin = append(varRightJoin, "cache", dbName, tableName)
		keyLeftJoin = append(keyLeftJoin, dbName, tableName)
	} else {
		varLeftJoin = append(varLeftJoin, "cache", tableName)
		varRightJoin = append(varRightJoin, "cache", tableName)
		keyLeftJoin = append(keyLeftJoin, tableName)
	}

	for _, field := range fields {
		varLeftJoin = append(varLeftJoin, field.Name.Source())
		varRightJoin = append(varRightJoin, field.Name.Source())
		keyLeftJoin = append(keyLeftJoin, field.Name.Source())
		keyRightJoin = append(keyRightJoin, util.EscapeGolangKeyword(stringx.From(field.Name.ToCamel()).UnTitle()))
		keyRightArgJoin = append(keyRightArgJoin, "%v")
		dataRightJoin = append(dataRightJoin, "data."+field.Name.ToCamel())
		fieldNameJoin = append(fieldNameJoin, field.Name.Source())
	}
	varLeftJoin = append(varLeftJoin, "prefix")
	keyLeftJoin = append(keyLeftJoin, "key")

	varLeft = util.SafeString(varLeftJoin.ToCamel().With("").UnTitle())
	varRight = fmt.Sprintf(`"%s"`, varRightJoin.ToCamel().UnTitle().With(":").Source()+":")
	varExpression = fmt.Sprintf(`%s = %s`, varLeft, varRight)

	keyLeft = util.SafeString(keyLeftJoin.ToCamel().With("").UnTitle())
	keyRight = fmt.Sprintf(`fmt.Sprintf("%s%s", %s, %s)`, "%s", keyRightArgJoin.With(":").Source(), varLeft, keyRightJoin.With(", ").Source())
	dataKeyRight = fmt.Sprintf(`fmt.Sprintf("%s%s", %s, %s)`, "%s", keyRightArgJoin.With(":").Source(), varLeft, dataRightJoin.With(", ").Source())
	keyExpression = fmt.Sprintf("%s := %s", keyLeft, keyRight)
	dataKeyExpression = fmt.Sprintf("%s := %s", keyLeft, dataKeyRight)

	return Key{
		VarLeft:           varLeft,
		VarRight:          varRight,
		VarExpression:     varExpression,
		KeyLeft:           keyLeft,
		KeyRight:          keyRight,
		DataKeyRight:      dataKeyRight,
		KeyExpression:     keyExpression,
		DataKeyExpression: dataKeyExpression,
		Fields:            fields,
		FieldNameJoin:     fieldNameJoin,
	}
}

// Title convert items into Title and return
func (j Join) Title() Join {
	var join Join
	for _, each := range j {
		join = append(join, stringx.From(each).Title())
	}

	return join
}

// ToCamel 转为驼峰。
func (j Join) ToCamel() Join {
	var join Join
	for _, each := range j {
		join = append(join, stringx.From(each).ToCamel())
	}
	return join
}

// ToSnake 转为蛇式。
func (j Join) ToSnake() Join {
	var join Join
	for _, each := range j {
		join = append(join, stringx.From(each).ToSnake())
	}

	return join
}

// UnTitle 首字母小写。
func (j Join) UnTitle() Join {
	var join Join
	for _, each := range j {
		join = append(join, stringx.From(each).UnTitle())
	}

	return join
}

// ToUpper 转为大写。
func (j Join) ToUpper() Join {
	var join Join
	for _, each := range j {
		join = append(join, stringx.From(each).ToUpper())
	}

	return join
}

// ToLower 转为小写。
func (j Join) ToLower() Join {
	var join Join
	for _, each := range j {
		join = append(join, stringx.From(each).ToLower())
	}

	return join
}

// With 使用分隔符连接。
func (j Join) With(sep string) stringx.String {
	return stringx.From(strings.Join(j, sep))
}
