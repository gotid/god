package pathvar

import (
	"context"
	"net/http"
)

var pathVars = contextKey("pathVars")

// Vars 解析路径变量并以字典形式返回。
func Vars(r *http.Request) map[string]string {
	vars, ok := r.Context().Value(pathVars).(map[string]string)
	if ok {
		return vars
	}

	return nil
}

// WithVars 写入参数至给定的请求并返回一个新的 *http.Request。
func WithVars(r *http.Request, params map[string]string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), pathVars, params))
}

type contextKey string

func (k contextKey) String() string {
	return "rest/pathvar/context key: " + string(k)
}
