package mapping

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gotid/god/lib/jsonx"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/stringx"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultKeyName = "key"
	delimiter      = '.'
)

var (
	errValueNotStruct   = errors.New("值类型不是 struct 结构体")
	errValueNotSettable = errors.New("值不可设置")
	errTypeMismatch     = errors.New("类型不匹配")

	keyUnmarshaler   = NewUnmarshaler(defaultKeyName)
	cacheKeys        = make(map[string][]string)
	cacheKeysLock    sync.Mutex
	durationType     = reflect.TypeOf(time.Duration(0))
	defaultCache     = make(map[string]any)
	defaultCacheLock sync.Mutex
	emptyMap         = map[string]any{}
	emptyValue       = reflect.ValueOf(lang.Placeholder)
)

type (
	// Unmarshaler 用于解组给定标签键。
	Unmarshaler struct {
		key  string
		opts unmarshalOptions
	}

	// UnmarshalOption 自定义一个解组器。
	UnmarshalOption func(options *unmarshalOptions)

	unmarshalOptions struct {
		fromString   bool
		canonicalKey func(key string) string
	}
)

// Unmarshal 解组字典 m 至 v。
func (u *Unmarshaler) Unmarshal(m map[string]any, v any) error {
	return u.UnmarshalValuer(mapValuer(m), v)
}

// UnmarshalValuer 解组字典 m 至 v。
func (u *Unmarshaler) UnmarshalValuer(m Valuer, v any) error {
	return u.unmarshalWithFullName(simpleValuer{current: m}, v, "")
}

func (u *Unmarshaler) unmarshalWithFullName(m valuerWithParent, v any, fullName string) error {
	rv := reflect.ValueOf(v)
	if err := ValidatePtr(&rv); err != nil {
		return err
	}

	rte := reflect.TypeOf(v).Elem()
	if rte.Kind() != reflect.Struct {
		return errValueNotStruct
	}

	rve := rv.Elem()
	numFields := rte.NumField()
	for i := 0; i < numFields; i++ {
		if err := u.processField(rte.Field(i), rve.Field(i), m, fullName); err != nil {
			return err
		}
	}

	return nil
}

func (u *Unmarshaler) processField(field reflect.StructField, value reflect.Value,
	m valuerWithParent, fullName string) error {
	if usingDifferentKeys(u.key, field) {
		return nil
	}

	if field.Anonymous {
		err := u.processAnonymousField(field, value, m, fullName)
		return err
	}

	err := u.processNamedField(field, value, m, fullName)
	return err
}

func (u *Unmarshaler) processAnonymousField(field reflect.StructField, value reflect.Value,
	m valuerWithParent, fullName string) error {
	key, options, err := u.parseOptionsWithContext(field, m, fullName)
	if err != nil {
		return err
	}

	if _, hasValue := getValue(m, key); hasValue {
		return fmt.Errorf("字段 %s 不能包裹在里面，因为它是匿名的", key)
	}

	if options.optional() {
		return u.processAnonymousFieldOptional(field, value, key, m, fullName)
	}

	return u.processAnonymousFieldRequired(field, value, m, fullName)
}

// 获取字典 m 中给定键 key 的值，键的格式可为 parentKey.childKey。
func getValue(m valuerWithParent, key string) (any, bool) {
	keys := readKeys(key)
	return getValueWithChainedKeys(m, keys)
}

func getValueWithChainedKeys(m valuerWithParent, keys []string) (any, bool) {
	switch len(keys) {
	case 0:
		return nil, false
	case 1:
		v, ok := m.Value(keys[0])
		return v, ok
	default:
		if v, ok := m.Value(keys[0]); ok {
			if nextM, ok := v.(map[string]any); ok {
				return getValueWithChainedKeys(recursiveValuer{
					current: mapValuer(nextM),
					parent:  m,
				}, keys[1:])
			}
		}

		return nil, false
	}
}

func readKeys(key string) []string {
	cacheKeysLock.Lock()
	keys, ok := cacheKeys[key]
	cacheKeysLock.Unlock()
	if ok {
		return keys
	}

	keys = strings.FieldsFunc(key, func(r rune) bool {
		return r == delimiter
	})
	cacheKeysLock.Lock()
	cacheKeys[key] = keys
	cacheKeysLock.Unlock()

	return keys
}

func (u *Unmarshaler) parseOptionsWithContext(field reflect.StructField, m Valuer, fullName string) (string, *fieldOptionsWithContext, error) {
	key, options, err := parseKeyAndOptions(u.key, field)
	if err != nil {
		return "", nil, err
	} else if options == nil {
		return key, nil, nil
	}

	optsWithContext, err := options.toOptionsWithContext(key, m, fullName)
	if err != nil {
		return "", nil, err
	}

	return key, optsWithContext, nil
}

func (u *Unmarshaler) processAnonymousFieldOptional(field reflect.StructField, value reflect.Value,
	key string, m valuerWithParent, fullName string) error {
	var filled bool
	var required int
	var requiredFilled int
	var indirectValue reflect.Value
	fieldType := Deref(field.Type)

	for i := 0; i < fieldType.NumField(); i++ {
		subField := fieldType.Field(i)
		fieldKey, fieldOpts, err := u.parseOptionsWithContext(subField, m, fullName)
		if err != nil {
			return err
		}

		_, hasValue := getValue(m, fieldKey)
		if hasValue {
			if !filled {
				filled = true
				maybeNewValue(field, value)
				indirectValue = reflect.Indirect(value)
			}
			if err = u.processField(subField, indirectValue.Field(i), m, fullName); err != nil {
				return err
			}
		}
		if !fieldOpts.optional() {
			required++
			if hasValue {
				requiredFilled++
			}
		}
	}

	if filled && required != requiredFilled {
		return fmt.Errorf("%s 未完全设置", key)
	}

	return nil
}

func (u *Unmarshaler) processAnonymousFieldRequired(field reflect.StructField, value reflect.Value,
	m valuerWithParent, fullName string) error {
	maybeNewValue(field, value)
	fieldType := Deref(field.Type)
	indirectValue := reflect.Indirect(value)

	for i := 0; i < fieldType.NumField(); i++ {
		if err := u.processField(fieldType.Field(i), indirectValue.Field(i), m, fullName); err != nil {
			return err
		}
	}

	return nil
}

func (u *Unmarshaler) processFieldWithEnvValue(field reflect.StructField, value reflect.Value,
	envVal string, opts *fieldOptionsWithContext, fullName string) error {
	fieldKind := field.Type.Kind()
	switch fieldKind {
	case reflect.Bool:
		val, err := strconv.ParseBool(envVal)
		if err != nil {
			return fmt.Errorf("用环境变量解组字段 %q 出错，%w", fullName, err)
		}

		value.SetBool(val)
		return nil
	case durationType.Kind():
		if err := fillDurationValue(fieldKind, value, envVal); err != nil {
			return fmt.Errorf("用环境变量解组字段 %q 出错，%w", fullName, err)
		}

		return nil
	case reflect.String:
		value.SetString(envVal)
		return nil
	default:
		return u.processFieldPrimitiveWithJSONNumber(field, value, json.Number(envVal), opts, fullName)
	}
}

func (u *Unmarshaler) processNamedField(field reflect.StructField, value reflect.Value,
	m valuerWithParent, fullName string) error {
	key, opts, err := u.parseOptionsWithContext(field, m, fullName)
	if err != nil {
		return err
	}

	fullName = join(fullName, key)
	if opts != nil && len(opts.EnvVar) > 0 {
		envVal := proc.Env(opts.EnvVar)
		if len(envVal) > 0 {
			return u.processFieldWithEnvValue(field, value, envVal, opts, fullName)
		}
	}

	canonicalKey := key
	if u.opts.canonicalKey != nil {
		canonicalKey = u.opts.canonicalKey(key)
	}

	valuer := createValuer(m, opts)
	mapValue, hasValue := getValue(valuer, canonicalKey)
	if !hasValue {
		return u.processNamedFieldWithoutValue(field, value, opts, fullName)
	}

	return u.processNamedFieldWithValue(field, value, valueWithParent{
		value:  mapValue,
		parent: valuer,
	}, key, opts, fullName)
}

func (u *Unmarshaler) processNamedFieldWithValue(field reflect.StructField, value reflect.Value,
	vp valueWithParent, key string, opts *fieldOptionsWithContext, fullName string) error {
	mapValue := vp.value
	if mapValue == nil {
		if opts.optional() {
			return nil
		}

		return fmt.Errorf("字段 %s 必须不为空", key)
	}

	if !value.CanSet() {
		return fmt.Errorf("field %s 不可设置", key)
	}

	maybeNewValue(field, value)

	if yes, err := u.processFieldTextUnmarshaler(field, value, mapValue); yes {
		return err
	}

	fieldKind := Deref(field.Type).Kind()
	switch fieldKind {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
		return u.processFieldNotFromString(field, value, vp, opts, fullName)
	default:
		if u.opts.fromString || opts.fromString() {
			valueKind := reflect.TypeOf(mapValue).Kind()
			if valueKind != reflect.String {
				return fmt.Errorf("错误：字典值的值不是字符串，而是 %s", valueKind)
			}

			options := opts.options()
			if len(options) > 0 {
				if !stringx.Contains(options, mapValue.(string)) {
					return fmt.Errorf(`错误：字段 "%s" 的值 "%s" 未定义在选项 "%v" 中`,
						key, vp, options)
				}
			}

			return fillPrimitive(field.Type, value, mapValue, opts, fullName)
		}

		return u.processFieldNotFromString(field, value, vp, opts, fullName)
	}
}

func (u *Unmarshaler) processFieldNotFromString(field reflect.StructField, value reflect.Value,
	vp valueWithParent, opts *fieldOptionsWithContext, fullName string) error {
	fieldType := field.Type
	derefedFieldType := Deref(fieldType)
	typeKind := derefedFieldType.Kind()
	valueKind := reflect.TypeOf(vp.value).Kind()
	mapValue := vp.value

	switch {
	case valueKind == reflect.Map && typeKind == reflect.Struct:
		if mv, ok := mapValue.(map[string]any); ok {
			return u.processFieldStruct(field, value, &simpleValuer{
				current: mapValuer(mv),
				parent:  vp.parent,
			}, fullName)
		} else {
			return errTypeMismatch
		}
	case valueKind == reflect.Map && typeKind == reflect.Map:
		return u.fillMap(field, value, mapValue)
	case valueKind == reflect.String && typeKind == reflect.Map:
		return u.fillMapFromString(value, mapValue)
	case valueKind == reflect.String && typeKind == reflect.Slice:
		return u.fillSliceFromString(fieldType, value, mapValue)
	case valueKind == reflect.String && derefedFieldType == durationType:
		return fillDurationValue(fieldType.Kind(), value, mapValue.(string))
	default:
		return u.processFieldPrimitive(field, value, mapValue, opts, fullName)
	}
}

func (u *Unmarshaler) processFieldPrimitive(field reflect.StructField, value reflect.Value,
	mapValue any, opts *fieldOptionsWithContext, fullName string) error {
	fieldType := field.Type
	typeKind := Deref(fieldType).Kind()
	valueKind := reflect.TypeOf(mapValue).Kind()

	switch {
	case typeKind == reflect.Slice && valueKind == reflect.Slice:
		return u.fillSlice(fieldType, value, mapValue)
	case typeKind == reflect.Map && valueKind == reflect.Map:
		return u.fillMap(field, value, mapValue)
	default:
		switch v := mapValue.(type) {
		case json.Number:
			return u.processFieldPrimitiveWithJSONNumber(field, value, v, opts, fullName)
		default:
			if typeKind == valueKind {
				if err := validateValueInOptions(mapValue, opts.options()); err != nil {
					return err
				}

				return fillWithSameType(field, value, mapValue, opts)
			}
		}
	}

	return newTypeMismatchError(fullName)
}

func (u *Unmarshaler) processFieldPrimitiveWithJSONNumber(field reflect.StructField, value reflect.Value,
	v json.Number, opts *fieldOptionsWithContext, fullName string) error {
	fieldType := field.Type
	fieldKind := fieldType.Kind()
	typeKind := Deref(fieldType).Kind()

	if err := validateJsonNumberRange(v, opts); err != nil {
		return err
	}

	if err := validateValueInOptions(v, opts.options()); err != nil {
		return err
	}

	if fieldKind == reflect.Ptr {
		value = value.Elem()
	}

	switch typeKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iValue, err := v.Int64()
		if err != nil {
			return err
		}

		value.SetInt(iValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		iValue, err := v.Int64()
		if err != nil {
			return err
		}

		if iValue < 0 {
			return fmt.Errorf("unmarshal %q with bad value %q", fullName, v.String())
		}

		value.SetUint(uint64(iValue))
	case reflect.Float32, reflect.Float64:
		fValue, err := v.Float64()
		if err != nil {
			return err
		}

		value.SetFloat(fValue)
	default:
		return newTypeMismatchError(fullName)
	}

	return nil
}
func (u *Unmarshaler) fillSliceFromString(fieldType reflect.Type, value reflect.Value,
	mapValue any) error {
	var slice []any
	switch v := mapValue.(type) {
	case fmt.Stringer:
		if err := jsonx.UnmarshalFromString(v.String(), &slice); err != nil {
			return err
		}
	case string:
		if err := jsonx.UnmarshalFromString(v, &slice); err != nil {
			return err
		}
	default:
		return errUnsupportedType
	}

	baseFieldType := Deref(fieldType.Elem())
	baseFieldKind := baseFieldType.Kind()
	conv := reflect.MakeSlice(reflect.SliceOf(baseFieldType), len(slice), cap(slice))

	for i := 0; i < len(slice); i++ {
		if err := u.fillSliceValue(conv, i, baseFieldKind, slice[i]); err != nil {
			return err
		}
	}

	value.Set(conv)
	return nil
}

func (u *Unmarshaler) generateMap(keyType, elemType reflect.Type, mapValue any) (reflect.Value, error) {
	mapType := reflect.MapOf(keyType, elemType)
	valueType := reflect.TypeOf(mapValue)
	if mapType == valueType {
		return reflect.ValueOf(mapValue), nil
	}

	refValue := reflect.ValueOf(mapValue)
	targetValue := reflect.MakeMapWithSize(mapType, refValue.Len())
	fieldElemKind := elemType.Kind()
	dereffedElemType := Deref(elemType)
	dereffedElemKind := dereffedElemType.Kind()

	for _, key := range refValue.MapKeys() {
		keythValue := refValue.MapIndex(key)
		keythData := keythValue.Interface()

		switch dereffedElemKind {
		case reflect.Slice:
			target := reflect.New(dereffedElemType)
			if err := u.fillSlice(elemType, target.Elem(), keythData); err != nil {
				return emptyValue, err
			}

			targetValue.SetMapIndex(key, target.Elem())
		case reflect.Struct:
			keythMap, ok := keythData.(map[string]any)
			if !ok {
				return emptyValue, errTypeMismatch
			}

			target := reflect.New(dereffedElemType)
			if err := u.Unmarshal(keythMap, target.Interface()); err != nil {
				return emptyValue, err
			}

			if fieldElemKind == reflect.Ptr {
				targetValue.SetMapIndex(key, target)
			} else {
				targetValue.SetMapIndex(key, target.Elem())
			}
		case reflect.Map:
			keythMap, ok := keythData.(map[string]any)
			if !ok {
				return emptyValue, errTypeMismatch
			}

			innerValue, err := u.generateMap(elemType.Key(), elemType.Elem(), keythMap)
			if err != nil {
				return emptyValue, err
			}

			targetValue.SetMapIndex(key, innerValue)
		default:
			switch v := keythData.(type) {
			case bool:
				targetValue.SetMapIndex(key, reflect.ValueOf(v))
			case string:
				targetValue.SetMapIndex(key, reflect.ValueOf(v))
			case json.Number:
				target := reflect.New(dereffedElemType)
				if err := setValue(dereffedElemKind, target.Elem(), v.String()); err != nil {
					return emptyValue, err
				}

				targetValue.SetMapIndex(key, target.Elem())
			default:
				targetValue.SetMapIndex(key, keythValue)
			}
		}
	}

	return targetValue, nil
}

func (u *Unmarshaler) fillMap(field reflect.StructField, value reflect.Value, mapValue any) error {
	if !value.CanSet() {
		return errValueNotSettable
	}

	fieldKeyType := field.Type.Key()
	fieldElemType := field.Type.Elem()
	targetValue, err := u.generateMap(fieldKeyType, fieldElemType, mapValue)
	if err != nil {
		return err
	}

	value.Set(targetValue)
	return nil
}

func (u *Unmarshaler) fillMapFromString(value reflect.Value, mapValue any) error {
	if !value.CanSet() {
		return errValueNotSettable
	}

	switch v := mapValue.(type) {
	case fmt.Stringer:
		if err := jsonx.UnmarshalFromString(v.String(), value.Addr().Interface()); err != nil {
			return err
		}
	case string:
		if err := jsonx.UnmarshalFromString(v, value.Addr().Interface()); err != nil {
			return err
		}
	default:
		return errUnsupportedType
	}

	return nil
}

func (u *Unmarshaler) fillSlice(fieldType reflect.Type, value reflect.Value, mapValue any) error {
	if !value.CanSet() {
		return errValueNotSettable
	}

	baseType := fieldType.Elem()
	baseKind := baseType.Kind()
	dereffedBaseType := Deref(baseType)
	dereffedBaseKind := dereffedBaseType.Kind()
	refValue := reflect.ValueOf(mapValue)
	if refValue.IsNil() {
		return nil
	}

	conv := reflect.MakeSlice(reflect.SliceOf(baseType), refValue.Len(), refValue.Cap())
	if refValue.Len() == 0 {
		value.Set(conv)
		return nil
	}

	var valid bool
	for i := 0; i < refValue.Len(); i++ {
		ithValue := refValue.Index(i).Interface()
		if ithValue == nil {
			continue
		}

		valid = true
		switch dereffedBaseKind {
		case reflect.Struct:
			target := reflect.New(dereffedBaseType)
			if err := u.Unmarshal(ithValue.(map[string]any), target.Interface()); err != nil {
				return err
			}

			if baseKind == reflect.Ptr {
				conv.Index(i).Set(target)
			} else {
				conv.Index(i).Set(target.Elem())
			}
		case reflect.Slice:
			if err := u.fillSlice(dereffedBaseType, conv.Index(i), ithValue); err != nil {
				return err
			}
		default:
			if err := u.fillSliceValue(conv, i, dereffedBaseKind, ithValue); err != nil {
				return err
			}
		}
	}

	if valid {
		value.Set(conv)
	}

	return nil
}

func (u *Unmarshaler) fillSliceValue(slice reflect.Value, index int,
	baseKind reflect.Kind, value any) error {
	ithVal := slice.Index(index)
	switch v := value.(type) {
	case fmt.Stringer:
		return setValue(baseKind, ithVal, v.String())
	case string:
		return setValue(baseKind, ithVal, v)
	default:
		// don't need to consider the difference between int, int8, int16, int32, int64,
		// uint, uint8, uint16, uint32, uint64, because they're handled as json.Number.
		if ithVal.Kind() == reflect.Ptr {
			baseType := Deref(ithVal.Type())
			if baseType.Kind() != reflect.TypeOf(value).Kind() {
				return errTypeMismatch
			}

			target := reflect.New(baseType).Elem()
			target.Set(reflect.ValueOf(value))
			ithVal.Set(target.Addr())
			return nil
		}

		if ithVal.Kind() != reflect.TypeOf(value).Kind() {
			return errTypeMismatch
		}

		ithVal.Set(reflect.ValueOf(value))
		return nil
	}
}

func (u *Unmarshaler) fillSliceWithDefault(derefedType reflect.Type, value reflect.Value,
	defaultValue string) error {
	baseFieldType := Deref(derefedType.Elem())
	baseFieldKind := baseFieldType.Kind()
	defaultCacheLock.Lock()
	slice, ok := defaultCache[defaultValue]
	defaultCacheLock.Unlock()
	if !ok {
		if baseFieldKind == reflect.String {
			slice = parseGroupedSegments(defaultValue)
		} else if err := jsonx.UnmarshalFromString(defaultValue, &slice); err != nil {
			return err
		}

		defaultCacheLock.Lock()
		defaultCache[defaultValue] = slice
		defaultCacheLock.Unlock()
	}

	return u.fillSlice(derefedType, value, slice)
}

func (u *Unmarshaler) processFieldTextUnmarshaler(field reflect.StructField, value reflect.Value,
	mapValue any) (bool, error) {
	var tval encoding.TextUnmarshaler
	var ok bool

	if field.Type.Kind() == reflect.Ptr {
		tval, ok = value.Interface().(encoding.TextUnmarshaler)
	} else {
		tval, ok = value.Addr().Interface().(encoding.TextUnmarshaler)
	}
	if ok {
		switch mv := mapValue.(type) {
		case string:
			return true, tval.UnmarshalText([]byte(mv))
		case []byte:
			return true, tval.UnmarshalText(mv)
		}
	}

	return false, nil
}

func (u *Unmarshaler) processNamedFieldWithoutValue(field reflect.StructField, value reflect.Value, opts *fieldOptionsWithContext, fullName string) error {
	derefType := Deref(field.Type)
	fieldKind := derefType.Kind()
	if defaultValue, ok := opts.getDefault(); ok {
		if field.Type.Kind() == reflect.Ptr {
			maybeNewValue(field, value)
			value = value.Elem()
		}
		if derefType == durationType {
			return fillDurationValue(fieldKind, value, defaultValue)
		}

		switch fieldKind {
		case reflect.Array, reflect.Slice:
			return u.fillSliceWithDefault(derefType, value, defaultValue)
		default:
			return setValue(fieldKind, value, defaultValue)
		}
	}

	switch fieldKind {
	case reflect.Array, reflect.Map, reflect.Slice:
		if !opts.optional() {
			return u.processFieldNotFromString(field, value, valueWithParent{
				value: emptyMap,
			}, opts, fullName)
		}
	case reflect.Struct:
		if !opts.optional() {
			required, err := structValueRequired(u.key, derefType)
			if err != nil {
				return err
			}

			if required {
				return fmt.Errorf("必填字段 %q 未设置", fullName)
			}

			return u.processFieldNotFromString(field, value, valueWithParent{
				value: emptyMap,
			}, opts, fullName)
		}
	default:
		if !opts.optional() {
			return newInitError(fullName)
		}
	}

	return nil
}

func (u *Unmarshaler) processFieldStruct(field reflect.StructField, value reflect.Value,
	m valuerWithParent, fullName string) error {
	if field.Type.Kind() == reflect.Ptr {
		baseType := Deref(field.Type)
		target := reflect.New(baseType).Elem()
		if err := u.unmarshalWithFullName(m, target.Addr().Interface(), fullName); err != nil {
			return err
		}

		value.Set(target.Addr())
	} else if err := u.unmarshalWithFullName(m, value.Addr().Interface(), fullName); err != nil {
		return err
	}

	return nil
}

func join(elem ...string) string {
	var builder strings.Builder

	var fillSep bool
	for _, e := range elem {
		if len(e) == 0 {
			continue
		}

		if fillSep {
			builder.WriteByte(delimiter)
		} else {
			fillSep = true
		}

		builder.WriteString(e)
	}

	return builder.String()
}

func fillDurationValue(fieldKind reflect.Kind, value reflect.Value, dur string) error {
	d, err := time.ParseDuration(dur)
	if err != nil {
		return err
	}

	if fieldKind == reflect.Ptr {
		value.Elem().Set(reflect.ValueOf(d))
	} else {
		value.Set(reflect.ValueOf(d))
	}

	return nil
}

func fillPrimitive(fieldType reflect.Type, value reflect.Value, mapValue any,
	opts *fieldOptionsWithContext, fullName string) error {
	if !value.CanSet() {
		return errValueNotSettable
	}

	baseType := Deref(fieldType)
	if fieldType.Kind() == reflect.Ptr {
		target := reflect.New(baseType).Elem()
		switch mapValue.(type) {
		case string, json.Number:
			value.Set(target.Addr())
			value = target
		}
	}

	switch v := mapValue.(type) {
	case string:
		return validateAndSetValue(baseType.Kind(), value, v, opts)
	case json.Number:
		if err := validateJsonNumberRange(v, opts); err != nil {
			return err
		}
		return setValue(baseType.Kind(), value, v.String())
	default:
		return newTypeMismatchError(fullName)
	}
}

func fillWithSameType(field reflect.StructField, value reflect.Value, mapValue any,
	opts *fieldOptionsWithContext) error {
	if !value.CanSet() {
		return errValueNotSettable
	}

	if err := validateValueRange(mapValue, opts); err != nil {
		return err
	}

	if field.Type.Kind() == reflect.Ptr {
		baseType := Deref(field.Type)
		target := reflect.New(baseType).Elem()
		setSameKindValue(baseType, target, mapValue)
		value.Set(target.Addr())
	} else {
		setSameKindValue(field.Type, value, mapValue)
	}

	return nil
}

func setSameKindValue(targetType reflect.Type, target reflect.Value, value any) {
	if reflect.ValueOf(value).Type().AssignableTo(targetType) {
		target.Set(reflect.ValueOf(value))
	} else {
		target.Set(reflect.ValueOf(value).Convert(targetType))
	}
}

func newInitError(name string) error {
	return fmt.Errorf("字段 %s 未设置", name)
}

func newTypeMismatchError(name string) error {
	return fmt.Errorf("错误：字段 %s 类型不匹配", name)
}

func createValuer(v valuerWithParent, opts *fieldOptionsWithContext) valuerWithParent {
	if opts.inherit() {
		return recursiveValuer{
			current: v,
			parent:  v.Parent(),
		}
	}

	return simpleValuer{
		current: v,
		parent:  v.Parent(),
	}
}

// NewUnmarshaler 返回一个 Unmarshaler。
func NewUnmarshaler(key string, opts ...UnmarshalOption) *Unmarshaler {
	unmarshaler := Unmarshaler{key: key}

	for _, opt := range opts {
		opt(&unmarshaler.opts)
	}

	return &unmarshaler
}

// UnmarshalKey 解组键值对字典 m 至 v。
func UnmarshalKey(m map[string]any, v any) error {
	return keyUnmarshaler.Unmarshal(m, v)
}

// WithStringValues 使用字符串形式的数值。
func WithStringValues() UnmarshalOption {
	return func(opt *unmarshalOptions) {
		opt.fromString = true
	}
}

// WithCanonicalKeyFunc 定义键的规范函数
func WithCanonicalKeyFunc(f func(string) string) UnmarshalOption {
	return func(opt *unmarshalOptions) {
		opt.canonicalKey = f
	}
}
