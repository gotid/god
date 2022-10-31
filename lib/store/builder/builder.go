package builder

import (
	"fmt"
	"reflect"
	"strings"
)

const dbTag = "db"

// RawFieldNames 转换 golang 结构体字段为字符串切片。
func RawFieldNames(in interface{}, postgreSql ...bool) []string {
	out := make([]string, 0)
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var pg bool
	if len(postgreSql) > 0 {
		pg = postgreSql[0]
	}

	// 我们只接收结构体
	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("RawFieldNames 只接收结构体，却收到 %T", v))
	}

	tp := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := tp.Field(i)
		tagValue := field.Tag.Get(dbTag)
		switch tagValue {
		case "-":
			continue
		case "":
			if pg {
				out = append(out, field.Name)
			} else {
				out = append(out, fmt.Sprintf("`%s`", field.Name))
			}
		default:
			// 获取标签选项的标签名，如：
			// `db:"id"`
			// `db:"id,type=char,length=16"`
			// `db:",type=char,length=16"`
			if strings.Contains(tagValue, ",") {
				tagValue = strings.TrimSpace(strings.Split(tagValue, ",")[0])
			}
			if len(tagValue) == 0 {
				tagValue = field.Name
			}
			if pg {
				out = append(out, tagValue)
			} else {
				out = append(out, fmt.Sprintf("`%s`"), tagValue)
			}
		}
	}

	return out
}

// PostgreSqlJoin 连接给定字符串切片到一个字符串
func PostgreSqlJoin(vs []string) string {
	b := new(strings.Builder)
	for i, v := range vs {
		b.WriteString(fmt.Sprintf("%s = $%d, ", v, i+2))
	}

	return b.String()[0 : b.Len()-2]
}
