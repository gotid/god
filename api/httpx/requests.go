package httpx

import (
	"github.com/gotid/god/api/internal/encoding"
	"github.com/gotid/god/api/internal/header"
	"github.com/gotid/god/api/pathvar"
	"github.com/gotid/god/lib/mapping"
	"io"
	"net/http"
	"strings"
)

const (
	pathKey           = "path"
	formKey           = "form"
	maxMemory         = 32 << 20 // 32MB
	maxBodyLen        = 8 << 20  // 8MB
	separator         = ";"
	tokensInAttribute = 2
)

var (
	pathUnmarshaler = mapping.NewUnmarshaler(pathKey, mapping.WithStringValues())
	formUnmarshaler = mapping.NewUnmarshaler(formKey, mapping.WithStringValues())
)

func Parse(r *http.Request, v interface{}) error {
	if err := ParsePath(r, v); err != nil {
		return err
	}

	if err := ParseForm(r, v); err != nil {
		return err
	}

	if err := ParseHeaders(r, v); err != nil {
		return err
	}

	return ParseJsonBody(r, v)
}

// ParseJsonBody 解析 post 请求的 json 正文中的键值对参数到 v。
// 默认只读取前 maxBodyLen 个字节。
func ParseJsonBody(r *http.Request, v interface{}) error {
	if withJsonBody(r) {
		reader := io.LimitReader(r.Body, maxBodyLen)
		return mapping.UnmarshalJsonReader(reader, v)
	}

	return mapping.UnmarshalJsonMap(nil, v)
}

// ParseHeaders 解析请求标头中的键值对参数到 v。
func ParseHeaders(r *http.Request, v interface{}) error {
	return encoding.ParseHeaders(r.Header, v)
}

// ParseHeader 解析请求头并以字典形式返回。
func ParseHeader(headerValue string) map[string]string {
	ret := make(map[string]string)
	fields := strings.Split(headerValue, separator)
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len(field) == 0 {
			continue
		}

		kv := strings.SplitN(field, "=", tokensInAttribute)
		if len(kv) != tokensInAttribute {
			continue
		}

		ret[kv[0]] = kv[1]
	}

	return ret
}

// ParseForm 解析表单请求中的键值对参数到 v。
func ParseForm(r *http.Request, v interface{}) error {
	params, err := GetFormValues(r)
	if err != nil {
		return err
	}

	return formUnmarshaler.Unmarshal(params, v)
}

// ParsePath 解析网址路径中的键值对参数到 v。
// 形如 https://localhost/tags/:tag
func ParsePath(r *http.Request, v interface{}) error {
	vars := pathvar.Vars(r)
	m := make(map[string]interface{}, len(vars))
	for k, v := range vars {
		m[k] = v
	}

	return pathUnmarshaler.Unmarshal(m, v)
}

// 根据请求头判断请求体是否为 json 格式
func withJsonBody(r *http.Request) bool {
	return r.ContentLength > 0 && strings.Contains(r.Header.Get(header.ContentType), header.ApplicationJson)
}
