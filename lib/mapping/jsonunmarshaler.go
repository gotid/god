package mapping

import (
	"io"

	"git.zc0901.com/go/god/lib/container/gmap"
	"git.zc0901.com/go/god/lib/jsonx"
)

const jsonTagKey = "json"

var jsonUnmarshaler = NewUnmarshaler(jsonTagKey)

func UnmarshalJsonBytes(content []byte, v interface{}) error {
	return unmarshalJsonBytes(content, v, jsonUnmarshaler)
}

func UnmarshalJsonReader(reader io.Reader) (*gmap.StrAnyMap, error) {
	return unmarshalJsonReader(reader)
}

func unmarshalJsonBytes(content []byte, v interface{}, unmarshaler *Unmarshaler) error {
	var m map[string]interface{}
	if err := jsonx.Unmarshal(content, &m); err != nil {
		return err
	}

	return unmarshaler.Unmarshal(m, v)
}

func unmarshalJsonReader(reader io.Reader) (*gmap.StrAnyMap, error) {
	var m map[string]interface{}
	if err := jsonx.UnmarshalFromReader(reader, &m); err != nil {
		return nil, err
	}

	// 弃用，改用gf方式以支持validator
	// return unmarshaler.Unmarshal(m, v)

	return gmap.NewStrAnyMapFrom(m), nil
}
