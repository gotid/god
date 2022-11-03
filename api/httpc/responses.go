package httpc

import (
	"bytes"
	"github.com/gotid/god/api/internal/encoding"
	"github.com/gotid/god/api/internal/header"
	"github.com/gotid/god/lib/mapping"
	"io"
	"net/http"
	"strings"
)

// Parse 解析响应。
func Parse(resp *http.Response, val interface{}) error {
	if err := ParseHeaders(resp, val); err != nil {
		return err
	}

	return ParseJsonBody(resp, val)
}

// ParseHeaders 解析响应头。
func ParseHeaders(resp *http.Response, val interface{}) error {
	return encoding.ParseHeaders(resp.Header, val)
}

// ParseJsonBody 解析 json 内容类型的响应体。
func ParseJsonBody(resp *http.Response, val interface{}) error {
	defer resp.Body.Close()

	if isContentTypeJson(resp) {
		if resp.ContentLength > 0 {
			return mapping.UnmarshalJsonReader(resp.Body, val)
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, resp.Body); err != nil {
			return err
		}

		if buf.Len() > 0 {
			return mapping.UnmarshalJsonReader(&buf, val)
		}
	}

	return mapping.UnmarshalJsonMap(nil, val)
}

func isContentTypeJson(resp *http.Response) bool {
	return strings.Contains(resp.Header.Get(header.ContentType), header.ApplicationJson)
}
