// Package gyaml provides accessing and converting for YAML content.
package gyaml

import (
	"github.com/gotid/god/internal/json"
	"github.com/gotid/god/lib/gconv"
	"gopkg.in/yaml.v3"
)

func Encode(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func Decode(v []byte) (interface{}, error) {
	var result map[string]interface{}
	if err := yaml.Unmarshal(v, &result); err != nil {
		return nil, err
	}
	return gconv.MapDeep(result), nil
}

func DecodeTo(v []byte, result interface{}) error {
	return yaml.Unmarshal(v, result)
}

func ToJson(v []byte) ([]byte, error) {
	if r, err := Decode(v); err != nil {
		return nil, err
	} else {
		return json.Marshal(r)
	}
}
