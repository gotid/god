package jsonx

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Marshal 编排 v 至字节切片。
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalToString 编排 v 至字符串。
func MarshalToString(v interface{}) (string, error) {
	data, err := Marshal(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Unmarshal 将 data 解编排至 v。
func Unmarshal(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := unmarshalUseNumber(decoder, v); err != nil {
		return formatError(string(data), err)
	}

	return nil
}

func unmarshalUseNumber(decoder *json.Decoder, v interface{}) error {
	decoder.UseNumber()
	return decoder.Decode(v)
}

func formatError(s string, err error) error {
	return fmt.Errorf("字符串：`%s`，错误：`%w`", s, err)
}
