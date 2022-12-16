package encoding

import (
	"bytes"
	"encoding/json"
	"github.com/gotid/god/lib/lang"
	"gopkg.in/yaml.v2"
)

func YamlToJson(data []byte) ([]byte, error) {
	var val any
	if err := yaml.Unmarshal(data, &val); err != nil {
		return nil, err
	}

	val = toStringKeyMap(val)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(val); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func toStringKeyMap(v any) any {
	switch v := v.(type) {
	case []any:
		return convertSlice(v)
	case map[any]any:
		return convertKeyToString(v)
	case bool, string:
		return v
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
		return convertNumberToJsonNumber(v)
	default:
		return lang.Repr(v)
	}
}

func convertNumberToJsonNumber(v any) json.Number {
	return json.Number(lang.Repr(v))
}

func convertKeyToString(m map[any]any) map[string]any {
	ret := make(map[string]any)
	for k, v := range m {
		ret[lang.Repr(k)] = toStringKeyMap(v)
	}

	return ret
}

func convertSlice(vs []any) any {
	ret := make([]any, len(vs))
	for i, v := range vs {
		ret[i] = toStringKeyMap(v)
	}

	return ret
}
