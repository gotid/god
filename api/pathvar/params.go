package pathvar

import (
	"context"
	"net/http"
)

// 路径变量键
var pathVars = contextKey("pathVars")

// Vars 解析路径变量并返回为映射。
func Vars(r *http.Request) map[string]string {
	vars, ok := r.Context().Value(pathVars).(map[string]string)
	if ok {
		return vars
	}

	return nil
}

// WithPathVars 将路径变量写入指定请求并返回新请求
func WithPathVars(r *http.Request, params map[string]string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), pathVars, params))
}

type contextKey string

func (c contextKey) String() string {
	return "api/internal/pathvar key: " + string(c)
}
