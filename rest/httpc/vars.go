package httpc

import "errors"

const (
	pathKey   = "path"
	formKey   = "form"
	headerKey = "header"
	jsonKey   = "json"
	slash     = "/"
	colon     = ':'
)

// ErrGetWithBody 代表 GET 请求错误的传递了请求体。
var ErrGetWithBody = errors.New("HTTP GET 不应该有请求体")
