package template

// Types 定义模型中用到的类型声明模板
const Types = `
type (
	{{.lowerStartCamelObject}}Model interface {
		{{.method}}
	}

	default{{.upperStartCamelObject}}Model struct {
		{{if .withCache}}sqlc.CachedConn{{else}}conn sqlx.Conn{{end}}
		table string
	}

	{{.upperStartCamelObject}} struct {
		{{.fields}}
	}
)
`
