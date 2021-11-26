package sqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/gconv"
	"git.zc0901.com/go/god/lib/stringx"

	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/mapping"
)

type UpdateArgs struct {
	Id     string
	Fields string
	Args   []interface{}
}

// ExtractUpdateArgs 返回提取后的SQL更新参数。
func ExtractUpdateArgs(allFieldList []string, updateMap g.Map) (*UpdateArgs, error) {
	vid, ok := updateMap["id"]
	id := gconv.Int64(vid)
	if !ok || id == 0 {
		return nil, errors.New("主键id必须传递")
	}

	var fields []string
	var args []interface{}

	for field, value := range updateMap {
		if field == "id" {
			continue
		}

		if !strings.HasPrefix(field, "`") {
			field = fmt.Sprintf("`%s`", field)
		}

		if stringx.Contains(allFieldList, field) {
			fields = append(fields, field)
			args = append(args, value)
		}
	}

	return &UpdateArgs{
		gconv.String(id),
		strings.Join(fields, "=?,") + "=?",
		args,
	}, nil
}

// format 格式查询字符串和参数
func format(query string, args ...interface{}) (string, error) {
	numArgs := len(args)
	if numArgs == 0 {
		return query, nil
	}

	var b strings.Builder
	var argIdx int
	bytes := len(query)

	for i := 0; i < bytes; i++ {
		ch := query[i]
		switch ch {
		case '?':
			if argIdx >= numArgs {
				return "", fmt.Errorf("错误：SQL中有 %d 个问号(?)，但提供的参数不够", argIdx)
			}

			writeValue(&b, args[argIdx])
			argIdx++
		case '$':
			var j int
			for j := i + 1; j < bytes; j++ {
				char := query[j]
				if char < '0' || char > '9' {
					break
				}
			}
			if j > i+1 {
				index, err := strconv.Atoi(query[i+1 : j])
				if err != nil {
					return "", err
				}

				if index > argIdx {
					argIdx = index
				}
				index--
				if index < 0 || index >= numArgs {
					return "", fmt.Errorf("错误：SQL index %d 越界", index)
				}

				writeValue(&b, args[index])
				i = j - 1
			}
		default:
			b.WriteByte(ch)
		}
	}

	if argIdx < numArgs {
		return "", fmt.Errorf("错误：提供了 %d 个参数，和SQL不匹配", argIdx)
	}

	return b.String(), nil
}

func writeValue(buf *strings.Builder, arg interface{}) {
	switch v := arg.(type) {
	case bool:
		if v {
			buf.WriteByte('1')
		} else {
			buf.WriteByte('0')
		}
	case string:
		buf.WriteByte('\'')
		buf.WriteString(escape(v))
		buf.WriteByte('\'')
	default:
		buf.WriteString(mapping.Repr(v))
	}
}

// escape 字符串转义
func escape(str string) string {
	var b strings.Builder

	for _, c := range str {
		switch c {
		case '\x00':
			b.WriteString(`\x00`)
		case '\r':
			b.WriteString(`\r`)
		case '\n':
			b.WriteString(`\n`)
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '"':
			b.WriteString(`\"`)
		case '\x1a':
			b.WriteString(`\x1a`)
		default:
			b.WriteRune(c)
		}
	}

	return b.String()
}

// 自动补全连接字符串
func prefectDSN(dataSourceName *string) {
	if strings.Count(*dataSourceName, "?") == 0 {
		*dataSourceName += "?"
	}

	var args []string
	if strings.Count(*dataSourceName, "parseTime=true") == 0 {
		args = append(args, "parseTime=true")
	}
	if strings.Count(*dataSourceName, "loc=Local") == 0 {
		args = append(args, "loc=Local")
	}
	if strings.HasSuffix(*dataSourceName, "?") {
		*dataSourceName += strings.Join(args, "&")
	} else {
		*dataSourceName += "&" + strings.Join(args, "&")
	}
}

func logSqlError(sql string, err error) {
	if err != nil && err != ErrNotFound {
		logx.Errorf("[SQL] %s >>> %s", err.Error(), sql)
	}
}

func logConnError(dsn string, err error) {
	dsn = desensitize(dsn)
	logx.Errorf("获取数据库实例失败 %s: %v", dsn, err)
}

// desensitize 脱敏：去除数据库连接中的账号密码
func desensitize(dsn string) string {
	pos := strings.LastIndex(dsn, "@")
	if pos >= 0 && pos+1 < len(dsn) {
		dsn = dsn[pos+1:]
	}
	return dsn
}

// scan 将数据库结果转换为Golang的数据类型
func scan(dest interface{}, rows *sql.Rows) error {
	// 验证接收目标必须为有效非空指针
	dv := reflect.ValueOf(dest)
	if err := mapping.ValidatePtr(&dv); err != nil {
		return err
	}

	// 将行数据扫描进目标结果
	dte := reflect.TypeOf(dest).Elem()
	dve := dv.Elem()
	switch dte.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		if dve.CanSet() {
			if !rows.Next() {
				if err := rows.Err(); err != nil {
					return err
				}
				return ErrNotFound
			}
			return rows.Scan(dest)
		} else {
			return ErrNotSettable
		}
	case reflect.Struct:
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return ErrNotFound
		}
		// 获取行的列名切片
		colNames, err := rows.Columns()
		if err != nil {
			return err
		}

		if values, err := mapStructFieldsIntoSlice(dve, colNames); err != nil {
			return err
		} else {
			return rows.Scan(values...)
		}
	case reflect.Slice:
		if !dve.CanSet() {
			return ErrNotSettable
		}

		ptr := dte.Elem().Kind() == reflect.Ptr
		appendFn := func(item reflect.Value) {
			if ptr {
				dve.Set(reflect.Append(dve, item))
			} else {
				dve.Set(reflect.Append(dve, reflect.Indirect(item)))
			}
		}
		fillFn := func(value interface{}) error {
			if dve.CanSet() {
				if err := rows.Scan(value); err != nil {
					return err
				} else {
					appendFn(reflect.ValueOf(value))
					return nil
				}
			}
			return ErrNotSettable
		}

		base := mapping.Deref(dte.Elem())
		switch base.Kind() {
		case reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			for rows.Next() {
				value := reflect.New(base)
				if err := fillFn(value.Interface()); err != nil {
					return err
				}
			}
		case reflect.Struct:
			// 获取行的列名切片
			colNames, err := rows.Columns()
			if err != nil {
				return err
			}

			for rows.Next() {
				value := reflect.New(base)
				if values, err := mapStructFieldsIntoSlice(value, colNames); err != nil {
					return err
				} else {
					if err := rows.Scan(values...); err != nil {
						return err
					} else {
						appendFn(value)
					}
				}
			}
		default:
			return ErrUnsupportedValueType
		}
		return nil
	default:
		return ErrUnsupportedValueType
	}
}

// 映射目标结构体字段到查询结果列，并赋初值
func mapStructFieldsIntoSlice(dve reflect.Value, columns []string) ([]interface{}, error) {
	columnValueMap, err := getFieldValueMap(dve)
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	if len(columnValueMap) != 0 {
		for i, column := range columns {
			if value, ok := columnValueMap[column]; ok {
				values[i] = value
			} else {
				var anonymous interface{}
				values[i] = &anonymous
			}
		}
	} else {
		fields := getFields(dve)

		for i := 0; i < len(values); i++ {
			field := fields[i]
			switch field.Kind() {
			case reflect.Ptr:
				if !field.CanInterface() {
					return nil, ErrNotReadableValue
				}
				if field.IsNil() {
					baseValueType := mapping.Deref(field.Type())
					field.Set(reflect.New(baseValueType))
				}
				values[i] = field.Interface()
			default:
				if !field.CanAddr() || !field.Addr().CanInterface() {
					return nil, ErrNotReadableValue
				}
				values[i] = field.Addr().Interface()
			}
		}
	}

	return values, nil
}

// getFieldValueMap: 获取结构体字段中标记的字段名——值映射关系
// 在编写字段tag的情况下，可以确保结构体字段和SQL选择列不一致的情况下不出错
func getFieldValueMap(dve reflect.Value) (map[string]interface{}, error) {
	t := mapping.Deref(dve.Type())
	size := t.NumField()
	result := make(map[string]interface{}, size)

	for i := 0; i < size; i++ {
		// 取字段标记中的列名，如`db:"total"` 中的 total
		fieldName := getFieldName(t.Field(i))
		if len(fieldName) == 0 {
			return nil, nil
		}

		// 读取指针字段或非指针字段的值
		field := reflect.Indirect(dve).Field(i)
		switch field.Kind() {
		case reflect.Ptr:
			if !field.CanInterface() {
				return nil, ErrNotReadableValue
			}
			if field.IsNil() {
				typ := mapping.Deref(field.Type())
				field.Set(reflect.New(typ))
			}
			result[fieldName] = field.Interface()
		default:
			if !field.CanAddr() || !field.Addr().CanInterface() {
				return nil, ErrNotReadableValue
			}
			result[fieldName] = field.Addr().Interface()
		}
	}

	return result, nil
}

// getFieldName 获取结构体字段中标记的中的数据库字段名
func getFieldName(field reflect.StructField) string {
	key := field.Tag.Get(tagName)
	if len(key) == 0 {
		return ""
	}
	return strings.Split(key, ",")[0]
}

// getFields 递归获取目标结构体的字段列表
func getFields(dve reflect.Value) []reflect.Value {
	var fields []reflect.Value
	v := reflect.Indirect(dve) // 指针取值

	// 递归目标字段
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// 指针取值
		if field.Kind() == reflect.Ptr && field.IsNil() {
			baseValueType := mapping.Deref(field.Type()) // 解引用，取值
			field.Set(reflect.New(baseValueType))
		}

		field = reflect.Indirect(field)
		structField := v.Type().Field(i)

		// 嵌套字段
		if field.Kind() == reflect.Struct && structField.Anonymous {
			fields = append(fields, getFields(field)...)
		} else {
			fields = append(fields, field)
		}
	}

	return fields
}

// In 构建一个长度为n的占位符
func In(n int) string {
	ps := make([]string, n)
	for i := 0; i < n; i++ {
		ps[i] = "?"
	}
	return strings.Join(ps, ",")
}
