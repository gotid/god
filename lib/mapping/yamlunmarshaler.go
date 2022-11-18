package mapping

import (
	"encoding/json"
	"errors"
	"io"

	"gopkg.in/yaml.v2"
)

// 为了使.json和.yaml保持一致，我们只使用json作为标记键。
const yamlTagKey = "json"

var (
	// ErrUnsupportedType 是一个表示配置格式不被支持的错误。
	ErrUnsupportedType = errors.New("只支持字典格式的配置")

	yamlUnmarshaler = NewUnmarshaler(yamlTagKey)
)

// UnmarshalYamlBytes 解编组 []byte 至 v。
func UnmarshalYamlBytes(content []byte, v any) error {
	return unmarshalYamlBytes(content, v, yamlUnmarshaler)
}

// UnmarshalYamlReader 解编组 io.Reader 至 v。
func UnmarshalYamlReader(reader io.Reader, v any) error {
	return unmarshalYamlReader(reader, v, yamlUnmarshaler)
}

func cleanupInterfaceMap(in map[any]any) map[string]any {
	res := make(map[string]any)
	for k, v := range in {
		res[Repr(k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupInterfaceNumber(in any) json.Number {
	return json.Number(Repr(in))
}

func cleanupInterfaceSlice(in []any) []any {
	res := make([]any, len(in))
	for i, v := range in {
		res[i] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v any) any {
	switch v := v.(type) {
	case []any:
		return cleanupInterfaceSlice(v)
	case map[any]any:
		return cleanupInterfaceMap(v)
	case bool, string:
		return v
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
		return cleanupInterfaceNumber(v)
	default:
		return Repr(v)
	}
}

func unmarshal(unmarshaler *Unmarshaler, o, v any) error {
	if m, ok := o.(map[string]any); ok {
		return unmarshaler.Unmarshal(m, v)
	}

	return ErrUnsupportedType
}

func unmarshalYamlBytes(content []byte, v any, unmarshaler *Unmarshaler) error {
	var o any
	if err := yamlUnmarshal(content, &o); err != nil {
		return err
	}

	return unmarshal(unmarshaler, o, v)
}

func unmarshalYamlReader(reader io.Reader, v any, unmarshaler *Unmarshaler) error {
	var res any
	if err := yaml.NewDecoder(reader).Decode(&res); err != nil {
		return err
	}

	return unmarshal(unmarshaler, cleanupMapValue(res), v)
}

// yamlUnmarshal yaml 转 map[string]any。
func yamlUnmarshal(in []byte, out any) error {
	var res any
	if err := yaml.Unmarshal(in, &res); err != nil {
		return err
	}

	*out.(*any) = cleanupMapValue(res)
	return nil
}
