package middleware

import "net/http"

type {{.name}} struct {
}

func New{{.name}}() *{{.name}} {
	return &{{.name}}{}
}

func (m *{{.name}})Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO 实现中间件函数并删除此行

		// 如果需要，传递到下一个处理程序
		next(w, r)
	}
}
