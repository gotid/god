package encoding

import (
	"github.com/gotid/god/lib/mapping"
	"net/http"
	"net/textproto"
)

const headerKey = "header"

var headerUnmarshaler = mapping.NewUnmarshaler(headerKey,
	mapping.WithStringValues(),
	mapping.WithCanonicalKeyFunc(textproto.CanonicalMIMEHeaderKey),
)

// ParseHeaders 解析 http 请求头。
func ParseHeaders(header http.Header, v interface{}) error {
	m := map[string]interface{}{}
	for k, v := range header {
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			m[k] = v
		}
	}

	return headerUnmarshaler.Unmarshal(m, v)
}
