package template

// New 定义用于创建模型示例的模板。
const New = `
func new{{.upperStartCamelObject}}Model(conn sqlx.Conn{{if .withCache}}, c cache.Config{{end}}) *default{{.upperStartCamelObject}}Model {
	return &default{{.upperStartCamelObject}}Model{
		{{if .withCache}}CachedConn: sqlc.NewConn(conn, c){{else}}conn:conn{{end}},
		table:      {{.table}},
	}
}
`
