package mapping

import (
	"io"

	"github.com/gotid/god/internal/encoding"
)

// UnmarshalYamlBytes 解编组 []byte 至 v。
func UnmarshalYamlBytes(content []byte, v any, opts ...UnmarshalOption) error {
	bs, err := encoding.YamlToJson(content)
	if err != nil {
		return err
	}

	return UnmarshalJsonBytes(bs, v, opts...)
}

// UnmarshalYamlReader 解编组 io.Reader 至 v。
func UnmarshalYamlReader(reader io.Reader, v any, opts ...UnmarshalOption) error {
	bs, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return UnmarshalYamlBytes(bs, v, opts...)
}
