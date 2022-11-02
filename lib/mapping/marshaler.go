package mapping

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	tagKVSeparator = ":"
	emptyTag       = ""
)

// Marshal 编组给定的值并字典形式返回。
// optional=another 并未实现，因为难以实现且不常用。
func Marshal(v interface{}) (map[string]map[string]interface{}, error) {
	ret := make(map[string]map[string]interface{})
	rt := reflect.TypeOf(v)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		value := rv.Field(i)
		if err := processMember(field, value, ret); err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func processMember(field reflect.StructField, value reflect.Value, ret map[string]map[string]interface{}) error {
	var key string
	var opt *fieldOptions
	var err error
	tag, ok := getTag(field)
	if !ok {
		tag = emptyTag
		key = field.Name
	} else {
		key, opt, err = parseKeyAndOptions(tag, field)
		if err != nil {
			return err
		}

		if err = validate(field, value, opt); err != nil {
			return err
		}
	}

	val := value.Interface()
	if opt != nil && opt.FromString {
		val = fmt.Sprint(val)
	}
	m, ok := ret[tag]
	if ok {
		m[key] = val
	} else {
		m = map[string]interface{}{
			key: val,
		}
	}
	ret[tag] = m

	return nil
}

func validate(field reflect.StructField, value reflect.Value, opt *fieldOptions) error {
	if opt == nil || !opt.Optional {
		if err := validateOptional(field, value); err != nil {
			return err
		}
	}

	if opt == nil {
		return nil
	}

	if opt.Optional && value.IsZero() {
		return nil
	}

	if len(opt.Options) > 0 {
		if err := validateOptions(value, opt); err != nil {
			return err
		}
	}

	if opt.Range != nil {
		if err := validateRange(value, opt); err != nil {
			return err
		}
	}

	return nil
}

// 验证值是否在区间内
func validateRange(value reflect.Value, opt *fieldOptions) error {
	var val float64
	switch v := value.Interface().(type) {
	case int:
		val = float64(v)
	case int8:
		val = float64(v)
	case int16:
		val = float64(v)
	case int32:
		val = float64(v)
	case int64:
		val = float64(v)
	case uint:
		val = float64(v)
	case uint8:
		val = float64(v)
	case uint16:
		val = float64(v)
	case uint32:
		val = float64(v)
	case uint64:
		val = float64(v)
	case float32:
		val = float64(v)
	case float64:
		val = v
	default:
		return fmt.Errorf("不支持的选项值类型 %q", value.Type().String())
	}

	if val < opt.Range.left ||
		(!opt.Range.leftInclude && val == opt.Range.left) ||
		val > opt.Range.right ||
		(!opt.Range.rightInclude && val == opt.Range.right) {
		return fmt.Errorf("选项值 %v 越界", value.Interface())
	}

	return nil
}

// 验证选项值是否在可选项中
func validateOptions(value reflect.Value, opt *fieldOptions) error {
	var found bool
	val := fmt.Sprint(value.Interface())
	for i := range opt.Options {
		if opt.Options[i] == val {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("选项值 %q 不在可选项中", val)
	}

	return nil
}

// 验证可选性
func validateOptional(field reflect.StructField, value reflect.Value) error {
	switch field.Type.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			return fmt.Errorf("字段 %q 不能为空", field.Name)
		}
	case reflect.Array, reflect.Slice, reflect.Map:
		if value.IsNil() || value.Len() == 0 {
			return fmt.Errorf("字段 %q 不能为空", field.Name)
		}
	}

	return nil
}

func getTag(field reflect.StructField) (string, bool) {
	tag := string(field.Tag)
	if i := strings.Index(tag, tagKVSeparator); i >= 0 {
		return strings.TrimSpace(tag[:i]), true
	}

	return strings.TrimSpace(tag), false
}
