package mapping

import (
	"github.com/gotid/god/lib/jsonx"
	"io"
)

const jsonTagKey = "json"

var jsonUnmarshaler = NewUnmarshaler(jsonTagKey)

// UnmarshalJsonBytes 解编组 []byte 至给定变量。
func UnmarshalJsonBytes(content []byte, v interface{}) error {
	return unmarshalJsonBytes(content, v, jsonUnmarshaler)
}

// UnmarshalJsonMap 解编组 map 至给定变量。
func UnmarshalJsonMap(m map[string]interface{}, v interface{}) error {
	return jsonUnmarshaler.Unmarshal(m, v)
}

// UnmarshalJsonReader 解编组 io.Reader 至给定变量。
func UnmarshalJsonReader(reader io.Reader, v interface{}) error {
	return unmarshalJsonReader(reader, v, jsonUnmarshaler)
}

func unmarshalJsonBytes(content []byte, v interface{}, unmarshaler *Unmarshaler) error {
	var m map[string]interface{}
	if err := jsonx.Unmarshal(content, &m); err != nil {
		return err
	}

	return unmarshaler.Unmarshal(m, v)
}

func unmarshalJsonReader(reader io.Reader, v interface{}, unmarshaler *Unmarshaler) error {
	var m map[string]interface{}
	if err := jsonx.UnmarshalFromReader(reader, &m); err != nil {
		return err
	}

	return unmarshaler.Unmarshal(m, v)
}
